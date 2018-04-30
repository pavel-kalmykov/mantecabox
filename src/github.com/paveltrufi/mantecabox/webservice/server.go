package webservice

import (
	"github.com/gin-gonic/gin"
)

func Server() *gin.Engine {
	r := gin.Default()

	r.GET("/users", GetUsers) // Useful?
	r.GET("/users/:username", GetUser)
	r.POST("/users", RegisterUser)
	r.PUT("/users/:username", ModifyUser)
	r.DELETE("/users/:username", DeleteUser)

	return r
}
