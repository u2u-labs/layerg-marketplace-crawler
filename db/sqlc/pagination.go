package db

import (
	"database/sql"
	"fmt"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"unicode"

	"github.com/Masterminds/squirrel"

	"github.com/gin-gonic/gin"
)

// Pagination holds the pagination information.
type Pagination[T any] struct {
	Page       int    `json:"page"`              // Current page number
	Limit      int    `json:"limit"`             // Number of items per page
	TotalItems int64  `json:"totalItems"`        // Total number of items available
	TotalPages int64  `json:"totalPages"`        // Total number of pages
	Holders    *int64 `json:"holders,omitempty"` // Optional holder field
	Data       []T    `json:"data"`              // The paginated items (can be any type)

}

type PaginationParams struct {
	Page   int
	Limit  int
	Offset int
}

// get limit and offset from query parameters
func GetLimitAndOffset(ctx *gin.Context) (int, int, int) {
	pageStr := ctx.Query("page")
	limitStr := ctx.Query("limit")

	// Default values
	pageNum := 1
	limitNum := 10

	// Parse page and limit
	if p, err := strconv.Atoi(pageStr); err == nil {
		pageNum = p
	}
	if l, err := strconv.Atoi(limitStr); err == nil {
		limitNum = l
	}

	// error if page is less than 1 or limit is greater than 100
	if pageNum < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Page number must be greater than 0"})
		return 0, 0, 0
	}
	if limitNum > 100 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Limit must be less than 100"})
		return 0, 0, 0
	}

	// Calculate offset
	offset := (pageNum - 1) * limitNum

	return pageNum, limitNum, offset
}

// Helper function to convert string to snake_case
func toSnakeCase(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && unicode.IsUpper(r) {
			result.WriteRune('_')
		}
		result.WriteRune(unicode.ToLower(r))
	}
	return result.String()
}

// QueryWithDynamicFilter retrieves a slice of items from the database
func QueryWithDynamicFilter[T any](db *sql.DB, tableName string, limit int, offset int, filterConditions map[string][]string) ([]T, error) {
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("*").From(tableName)

	// Apply dynamic filters
	for column, values := range filterConditions {
		if len(values) == 1 {
			queryBuilder = queryBuilder.Where(squirrel.Eq{column: values[0]})
		} else if len(values) > 1 {
			queryBuilder = queryBuilder.Where(squirrel.Eq{column: values})
		}
	}

	// Apply pagination
	if limit > 0 {
		queryBuilder = queryBuilder.Limit(uint64(limit))
	}
	if offset > 0 {
		queryBuilder = queryBuilder.Offset(uint64(offset))
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("error building SQL query: %w", err)
	}

	rows, err := db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("error executing query: %w", err)
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("error getting columns: %w", err)
	}

	var items []T
	var item T
	itemType := reflect.TypeOf(item)
	if itemType.Kind() == reflect.Ptr {
		itemType = itemType.Elem()
	}

	// Create a map of DB column names to struct field indices
	columnMap := make(map[string]int)
	for i := 0; i < itemType.NumField(); i++ {
		field := itemType.Field(i)

		// Check for db tag first
		dbTag := field.Tag.Get("db")
		if dbTag != "" {
			columnMap[dbTag] = i
			continue
		}

		// Then check for json tag
		jsonTag := field.Tag.Get("json")
		if jsonTag != "" {
			// Split the json tag to handle options like omitempty
			tagParts := strings.Split(jsonTag, ",")
			columnMap[toSnakeCase(tagParts[0])] = i
			continue
		}

		// If no tags present, use field name in snake_case
		columnMap[toSnakeCase(field.Name)] = i
	}

	for rows.Next() {
		itemValue := reflect.New(itemType).Elem()
		scanArgs := make([]interface{}, len(columns))

		for i, colName := range columns {
			if fieldIndex, ok := columnMap[colName]; ok {
				field := itemValue.Field(fieldIndex)

				// Handle null values for pointer fields
				if field.Kind() == reflect.Ptr {
					if field.IsNil() {
						field.Set(reflect.New(field.Type().Elem()))
					}
					scanArgs[i] = field.Interface()
				} else {
					scanArgs[i] = field.Addr().Interface()
				}
			} else {
				var placeholder interface{}
				scanArgs[i] = &placeholder
			}
		}

		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("error scanning row: %w", err)
		}

		items = append(items, itemValue.Interface().(T))
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return items, nil
}

// CountItems counts the number of items in the database based on dynamic filters.
func CountItemsWithFilter(db *sql.DB, tableName string, filterConditions map[string][]string) (int, int64, error) {
	// Create a Squirrel query builder for counting items
	psql := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	queryBuilder := psql.Select("COUNT(*)").From(tableName)

	// Create a Squirrel query builder for counting items holder
	psql2 := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar)
	holderQueryBuilder := psql2.Select("COUNT(DISTINCT(owner))").From(tableName)

	// Apply dynamic filters
	for column, values := range filterConditions {
		if len(values) == 1 {
			queryBuilder = queryBuilder.Where(squirrel.Eq{column: values[0]})
			holderQueryBuilder = holderQueryBuilder.Where(squirrel.Eq{column: values[0]})
		} else if len(values) > 1 {
			queryBuilder = queryBuilder.Where(squirrel.Eq{column: values})
			holderQueryBuilder = holderQueryBuilder.Where(squirrel.Eq{column: values})
		}
	}

	// Convert the query to SQL
	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return 0, 0, err
	}

	holderQuery, holderArgs, err := holderQueryBuilder.ToSql()
	if err != nil {
		return 0, 0, err
	}

	// Execute the query
	var itemCount int
	var holderCount int64

	err = db.QueryRow(query, args...).Scan(&itemCount)
	if err != nil {
		return 0, 0, err
	}

	err = db.QueryRow(holderQuery, holderArgs...).Scan(&holderCount)
	if err != nil {
		return 0, 0, err
	}

	return itemCount, holderCount, nil
}
