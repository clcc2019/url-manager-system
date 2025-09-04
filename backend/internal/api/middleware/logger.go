package middleware

import (
	"time"

	"github.com/gin-contrib/logger"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// SetupLogger 设置日志中间件
func SetupLogger() gin.HandlerFunc {
	return logger.SetLogger(
		logger.WithLogger(func(c *gin.Context, latency time.Duration, clientIP, method, path string, statusCode int, bodySize int, userAgent string) {
			logrus.WithFields(logrus.Fields{
				"client_ip":   clientIP,
				"method":      method,
				"path":        path,
				"status_code": statusCode,
				"latency":     latency,
				"body_size":   bodySize,
				"user_agent":  userAgent,
			}).Info("HTTP request")
		}),
	)
}
