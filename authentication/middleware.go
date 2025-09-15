package authentication

import (
	"main/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

func Middleware(c *gin.Context) {
	token := c.Request.Header.Get("Authorization")

	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{utils.Err: "authorization"})
		c.Abort()
		return
	}

	parts := strings.Split(token, " ")
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		c.JSON(http.StatusUnauthorized, gin.H{utils.Err: "invalid authorized"})
		c.Abort()
		return
	}

	tokenStr := parts[1]

	if err := utils.VerifyToken(tokenStr); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{utils.Err: "Invalid or expired token"})
		c.Abort()
		return
	}

	c.Set("user", tokenStr)
	c.Next()
}
