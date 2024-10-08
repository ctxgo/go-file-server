package middlewares

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method

		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k := range c.Request.Header {
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			//c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Origin", origin)

			//c.Header("Access-Control-Allow-Origin", "*")
			c.Header("Access-Control-Allow-Headers", headerStr)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT,  Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma,X-BF-NONCE,X-BF-TIMESTAMP,X-BF-DEVICE-ID,X-BF-PLATFORM,X-BF-APP-VERSION,X-BF-APP-CHANNEL-CODE,X-BF-APP-NETWORK,X-BF-MODEL-NAME,X-BF-SYSTEM-VERSION,X-BF-APP-SCREENWIDTH,X-BF-APP-SCREENHEIGHT,X-BF-APP-BUILD")
			c.Header("Access-Control-Expose-Headers", "Authorization, Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Content-Type,X-BF-NONCE,X-BF-TIMESTAMP,X-BF-DEVICE-ID,X-BF-PLATFORM,X-BF-APP-VERSION,X-BF-APP-CHANNEL-CODE,X-BF-APP-NETWORK,X-BF-MODEL-NAME,X-BF-SYSTEM-VERSION,X-BF-APP-SCREENWIDTH,X-BF-APP-SCREENHEIGHT,X-BF-APP-BUILD")
			// c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
			//c.Set("content-type", "application/json")
		}

		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
	}
}
