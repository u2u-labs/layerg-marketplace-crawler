package response

import "github.com/gin-gonic/gin"

type ResponseData struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type ErrorResponse struct {
	Error  string `json:"error"`
	Detail string `json:"detail"`
}

// success response data
func SuccessReponseData(c *gin.Context, httpCode int, data interface{}) {
	c.JSON(httpCode, ResponseData{
		Message: "success",
		Data:    data,
	})

}

// error response data
func ErrorResponseData(c *gin.Context, httpCode int, message string) {
	c.JSON(httpCode, ResponseData{
		Message: message,
		Data:    nil,
	})
}
