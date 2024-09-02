package utils

import "github.com/gin-gonic/gin"

func GetHost(c *gin.Context) string {
	host := c.GetHeader("x-forwarded-host")
	if host != "" {
		return host
	}
	return c.GetHeader("Origin")
}
