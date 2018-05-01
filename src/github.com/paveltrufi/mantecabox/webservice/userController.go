package webservice

import (
	"database/sql"
	"net/http"

	"github.com/PeteProgrammer/go-automapper"
	"github.com/gin-gonic/gin"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/services"
)

func GetUsers(c *gin.Context) {
	users, err := services.GetUsers()
	if err != nil {
		sendJsonError(c, http.StatusInternalServerError, "Unable to retrieve users: "+err.Error())
		return
	}
	var dtos []models.UserDto
	automapper.Map(users, &dtos)
	c.JSON(http.StatusOK, dtos)
}

func GetUser(c *gin.Context) {
	username := c.Param("username")
	user, err := services.GetUser(username)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonError(c, http.StatusNotFound, "Unable to find user: "+username)
		} else {
			sendJsonError(c, http.StatusInternalServerError, "Unable to find user: "+err.Error())
		}
		return
	}
	var dto models.UserDto
	automapper.Map(user, &dto)
	c.JSON(http.StatusOK, dto)
}

func RegisterUser(c *gin.Context) {
	var credentials models.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		sendJsonError(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		return
	}
	registeredUser, err := services.RegisterUser(&credentials)
	if err != nil {
		sendJsonError(c, http.StatusBadRequest, "Unable to register user: "+err.Error())
		return
	}
	var dto models.UserDto
	automapper.Map(registeredUser, &dto)
	c.JSON(http.StatusCreated, dto)
}

func ModifyUser(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		sendJsonError(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		return
	}
	username := c.Param("username")
	user, err = services.ModifyUser(username, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonError(c, http.StatusNotFound, "Unable to find user: "+username)
		} else {
			sendJsonError(c, http.StatusBadRequest, "Unable to modify user: "+err.Error())
		}
		return
	}
	var dto models.UserDto
	automapper.Map(user, &dto)
	c.JSON(http.StatusCreated, dto)
}

func DeleteUser(c *gin.Context) {
	username := c.Param("username")
	err := services.DeleteUser(username)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonError(c, http.StatusNotFound, "Unable to find user: "+username)
		} else {
			sendJsonError(c, http.StatusBadRequest, "Unable to delete user: "+err.Error())
		}
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func sendJsonError(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, models.ServerError{
		Message: msg,
	})
}
