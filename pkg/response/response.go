package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
	Meta    *Meta       `json:"meta,omitempty"`
}

func Success(c *gin.Context, message string, data interface{}, meta *Meta) {
	c.JSON(http.StatusOK, Response{Message: message, Data: data, Meta: meta})
}

func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, Response{Message: "created", Data: data, Meta: nil})
}

func BadRequest(c *gin.Context, message string) {
	c.JSON(http.StatusBadRequest, Response{Message: message, Data: nil, Meta: nil})
}

func InternalServerError(c *gin.Context, message string) {
	c.JSON(http.StatusInternalServerError, Response{Message: message, Data: nil, Meta: nil})
}

func NotFound(c *gin.Context, message string) {
	c.JSON(http.StatusNotFound, Response{Message: message, Data: nil, Meta: nil})
}

func Unauthorized(c *gin.Context, message string) {
	c.JSON(http.StatusUnauthorized, Response{Message: message, Data: nil, Meta: nil})
}

func Forbidden(c *gin.Context, message string) {
	c.JSON(http.StatusForbidden, Response{Message: message, Data: nil, Meta: nil})
}

func BadGateway(c *gin.Context, message string) {
	c.JSON(http.StatusBadGateway, Response{Message: message, Data: nil, Meta: nil})
}

func GatewayTimeout(c *gin.Context, message string) {
	c.JSON(http.StatusGatewayTimeout, Response{Message: message, Data: nil, Meta: nil})
}

func ServiceUnavailable(c *gin.Context, message string) {
	c.JSON(http.StatusServiceUnavailable, Response{Message: message, Data: nil, Meta: nil})
}

func TooManyRequests(c *gin.Context, message string) {
	c.JSON(http.StatusTooManyRequests, Response{Message: message, Data: nil, Meta: nil})
}
