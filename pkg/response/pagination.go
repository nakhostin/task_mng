package response

import (
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

type Pagination struct {
	Page  int    `query:"page"`
	Limit int    `query:"limit"`
	Sort  string `query:"sort"`
}

func NewPagination(c *gin.Context) *Pagination {
	page := 1
	if p := c.Query("page"); p != "" {
		if val, err := strconv.Atoi(p); err == nil && val > 0 {
			page = val
		}
	}

	limit := 10
	if l := c.Query("limit"); l != "" {
		if val, err := strconv.Atoi(l); err == nil && val > 0 {
			limit = val
		}
	}

	sort := c.DefaultQuery("sort", "-created_at")

	// Parse sort string: "-field" -> "field DESC", "field" -> "field ASC"
	parsedSort := parseSort(sort)

	return &Pagination{Page: page, Limit: limit, Sort: parsedSort}
}

// parseSort converts API sort format to SQL ORDER BY format
// Examples: "-created_at" -> "created_at DESC", "name" -> "name ASC"
func parseSort(sort string) string {
	if sort == "" {
		return "created_at DESC"
	}

	// Check if sort starts with minus sign (descending order)
	if strings.HasPrefix(sort, "-") {
		fieldName := strings.TrimPrefix(sort, "-")
		return fieldName + " DESC"
	}

	// Default to ascending order
	return sort + " ASC"
}
