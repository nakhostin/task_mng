package response

import (
	"github.com/asaskevich/govalidator"
	"github.com/gin-gonic/gin"
)

func Parse[T any](c *gin.Context) (*T, error) {
	var form T
	if err := c.ShouldBindJSON(&form); err != nil {
		return nil, err
	}

	if _, err := govalidator.ValidateStruct(form); err != nil {
		return nil, err
	}

	return &form, nil
}

func ParseQuery[T any](c *gin.Context) (*T, error) {
	var form T
	if err := c.ShouldBindQuery(&form); err != nil {
		return nil, err
	}

	return &form, nil
}
