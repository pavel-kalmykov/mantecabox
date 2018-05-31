package webservice

import (
	"github.com/gin-gonic/gin"
)

func Router(useJWT bool) *gin.Engine {
	r := gin.Default()

	r.POST("/register", RegisterUser)
	r.POST("/2fa-verification", Generate2FAAndSendMail)
	r.POST("/login", AuthMiddleware.LoginHandler)
	r.GET("/refresh-token", AuthMiddleware.RefreshHandler)

	users := r.Group("/users")
	if useJWT {
		users.Use(AuthMiddleware.MiddlewareFunc())
	}

	users.GET("", GetUsers) // Useful?
	users.GET("/:email", GetUser)
	users.PUT("/:email", ModifyUser)
	users.DELETE("/:email", DeleteUser)


	files := r.Group("/files")
	if useJWT {
		files.Use(AuthMiddleware.MiddlewareFunc())
	}

	files.GET("/:file", GetFile)
	files.POST("", UploadFile)
	files.DELETE("/:file", DeleteFile)

	return r
}
