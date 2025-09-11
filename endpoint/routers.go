// Package endpoint
package endpoint

import (
	"main/authentication"
	"main/router"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AuthRouters(r *gin.Engine) {
	routers := r.Group("/")
	{
		routers.POST("register", authentication.Register)
		routers.POST("login", authentication.LoginUser)
		routers.POST("logout", authentication.Logout)
		routers.POST("contactFriends", authentication.ContactFriends)
		routers.POST("getProfile", authentication.GetProfile)
		routers.POST("deleteUserProfile", authentication.DeleteUserProfile)
		routers.POST("deleteUserMain", authentication.DeleteUserMain)
	}
}

func HandlerRouter(r *gin.Engine) {
	routers := r.Group("/ws")
	{
		routers.GET("/addCountact", router.AddCountact)
		routers.GET("/sendMessage", router.SendMessage)
		routers.GET("/getMessage", router.GetMessage)
	}

}

func ProtectedRouters(r *gin.Engine) {
	protected := r.Group("/protected")
	{
		protected.Use(authentication.Middleware).GET("/profile", func(ctx *gin.Context) {
			user, _ := ctx.Get("user")
			ctx.JSON(http.StatusOK, gin.H{
				"message": "welcome to protected",
				"user":    user,
			})
		})
	}
}
