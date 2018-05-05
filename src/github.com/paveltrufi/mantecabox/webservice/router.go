package webservice

import (
	"github.com/gin-gonic/gin"
)

func Router(userJWT bool) *gin.Engine {
	r := gin.Default()

	r.POST("/register", RegisterUser)
	r.POST("/login", AuthMiddleware.LoginHandler)
	r.GET("/refresh_token", AuthMiddleware.RefreshHandler)

	users := r.Group("/users")
	if userJWT {
		users.Use(AuthMiddleware.MiddlewareFunc())
	}
	users.GET("", GetUsers) // Useful?
	users.GET("/:username", GetUser)
	users.PUT("/:username", ModifyUser)
	users.DELETE("/:username", DeleteUser)

	return r
}
