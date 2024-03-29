package middlewares

import (
	"crypto/rand"
	"fmt"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
)

func InitializeSession(app *gin.Engine) {
	var key []byte = make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		panic(err)
	}
	store := cookie.NewStore([]byte("session"), key)
	//store.Options
	app.Use(sessions.Sessions("user-session", store))
}

func ValidateAuthentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !isAuthenticated(c) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"Error:": "Connection is not authorized."})
		}
		c.Next()
	}

}

func isAuthenticated(c *gin.Context) bool {
	session := sessions.Default(c)
	sessionId := session.Get(c.Request.Header.Get("Authorization"))
	fmt.Println(c.Request.Header.Get("Authorization"))
	fmt.Println(sessionId)
	return sessionId == c.Request.Header.Get("Session")
}
