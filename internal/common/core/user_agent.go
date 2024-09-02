package core

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/mssola/user_agent"
)

func GetUserAgent(c *gin.Context) *user_agent.UserAgent {
	userAgentString := c.Request.UserAgent()
	return user_agent.New(userAgentString)
}

func FormatBrowserInfo(ua *user_agent.UserAgent) string {
	browser, version := ua.Browser()
	return fmt.Sprintf("Browser: %s, Version: %s", browser, version)
}
