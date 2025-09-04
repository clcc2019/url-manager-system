package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Recovery 自定义恢复中间件
func Recovery() gin.HandlerFunc {
	return gin.CustomRecovery(func(c *gin.Context, recovered interface{}) {
		if err, ok := recovered.(string); ok {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "Internal server error",
				"details": err,
			})
		}
		c.Abort()
	})
}
