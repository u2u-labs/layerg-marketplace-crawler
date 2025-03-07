package db

import "database/sql"

type DBManager struct {
	CrawlerDB     *sql.DB
	MarketplaceDB *sql.DB
	CrQueries     *Queries
	MpQueries     *Queries
}

func NewDBManager(crawlerConn, marketplaceConn *sql.DB) *DBManager {
	return &DBManager{
		CrawlerDB:     crawlerConn,
		MarketplaceDB: marketplaceConn,
		CrQueries:     New(crawlerConn),
		MpQueries:     New(marketplaceConn),
	}
}
