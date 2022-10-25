package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	token "github.com/mauroarnedo/ecommerce/tokens"
)

func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := c.Request.Header.Get("token")
		if clientToken == "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": "No authorization header provided"})
			c.Abort()
			return
		}
		claims, err := token.ValidateToken(clientToken)
		if err != "" {
			c.JSON(http.StatusInternalServerError, gin.H{"Error": err})
			c.Abort()
			return
		}
		c.Set("email", claims.Email)
		c.Set("uid", claims.Uid)
		c.Next()
	}
}
