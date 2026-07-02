package response

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type BaseResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func Success(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, BaseResponse{Success: true, Data: data})
}
func Fail(c *gin.Context, status int, code, message string) {
	c.JSON(status, BaseResponse{Success: false, Error: &APIError{Code: code, Message: message}})
}
