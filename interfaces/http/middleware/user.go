package middleware

import (
	"strconv"
	"task_mng/pkg/jwt"
	"task_mng/pkg/response"

	"github.com/gin-gonic/gin"
)

func LoginRequired(jwtManager *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "not_logged_in")
			c.Abort()
			return
		}

		if len(authHeader) < 7 || authHeader[:7] != "Bearer " {
			response.Unauthorized(c, "invalid_authorization_header_format")
			c.Abort()
			return
		}

		token := authHeader[7:]

		claims, err := jwtManager.ValidateAccessToken(token)
		if err != nil {
			response.Unauthorized(c, "not_logged_in")
			c.Abort()
			return
		}

		userID, err := strconv.ParseUint(claims.UserID, 10, 32)
		if err != nil {
			response.Unauthorized(c, "invalid_user_id")
			c.Abort()
			return
		}

		c.Set("user_id", uint(userID))

		c.Next()
	}
}
