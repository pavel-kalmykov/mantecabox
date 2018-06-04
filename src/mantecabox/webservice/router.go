package webservice

import (
	"mantecabox/models"

	"github.com/gin-gonic/gin"
	"github.com/zsais/go-gin-prometheus"
)

func Router(useJWT bool, configuration *models.Configuration) *gin.Engine {
	userController := NewUserController(configuration)
	if userController == nil {
		return nil
	}
	fileController := NewFileController(configuration)
	if fileController == nil {
		return nil
	}

	r := gin.Default()
	p := ginprometheus.NewPrometheus("gin")
	p.Use(r)

	r.POST("/register", userController.RegisterUser)
	r.POST("/2fa-verification", userController.Generate2FAAndSendMail)
	r.POST("/login", userController.AuthMiddleware().LoginHandler)
	r.GET("/refresh-token", userController.AuthMiddleware().RefreshHandler)

	users := r.Group("/users")
	if useJWT {
		users.Use(userController.AuthMiddleware().MiddlewareFunc())
	}

	users.GET("", userController.GetUsers) // Useful?
	users.GET("/:email", userController.GetUser)
	users.PUT("/:email", userController.ModifyUser)
	users.DELETE("/:email", userController.DeleteUser)

	files := r.Group("/files")
	if useJWT {
		files.Use(userController.AuthMiddleware().MiddlewareFunc())
	}

	files.GET("/:file", fileController.GetFile)
	files.GET("/:file/download", fileController.DownloadFile)

	files.GET("/:file/versions", fileController.GetAllFileVersions)
	files.GET("/:file/versions/:version", fileController.GetFileVersion)
	files.GET("/:file/versions/:version/download", fileController.DownloadFileVersion)

	files.GET("", fileController.GetAllFiles)
	files.POST("", fileController.UploadFile)
	files.DELETE("/:file", fileController.DeleteFile)

	return r
}
