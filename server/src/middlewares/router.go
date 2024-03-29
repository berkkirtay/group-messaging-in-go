package middlewares

import (
	"main/controllers"
	"net/http"

	"github.com/gin-gonic/gin"
)

func InitializeRouters(routerGroup *gin.RouterGroup) {
	routerGroup.Use(gin.CustomRecovery(handleGenericPanic))
	controllers.UserRouter(routerGroup)
	controllers.AuthRouter(routerGroup)
	routerGroup.Use(ValidateAuthentication())
	controllers.AdministrationRouter(routerGroup)
	controllers.CommandRouter(routerGroup)
	controllers.Roomouter(routerGroup)

}

func handleGenericPanic(c *gin.Context, err any) {
	defer func() {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"Error:": err})
	}()
}
