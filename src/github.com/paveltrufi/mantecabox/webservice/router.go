package webservice

import (
	"github.com/gin-gonic/gin"
)

func Router(useJWT bool) *gin.Engine {
	r := gin.Default()

	r.POST("/register", RegisterUser)
	r.POST("/login", AuthMiddleware.LoginHandler)
	r.GET("/refresh_token", AuthMiddleware.RefreshHandler)

	users := r.Group("/users")
	if useJWT {
		users.Use(AuthMiddleware.MiddlewareFunc())
	}

	users.GET("", GetUsers) // Useful?
	users.GET("/:username", GetUser)
	users.PUT("/:username", ModifyUser)
	users.DELETE("/:username", DeleteUser)

	// TODO: include within the user session in the future
	r.POST("/files", UploadFile)

	return r
}
