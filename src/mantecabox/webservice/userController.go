package webservice

import (
	"database/sql"
	"net/http"
	"time"

	"mantecabox/models"
	"mantecabox/services"
	"mantecabox/utilities/aes"

	"github.com/PeteProgrammer/go-automapper"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
)

// the jwt middleware
var AuthMiddleware = &jwt.GinJWTMiddleware{
	Realm:      "Mantecabox",
	Key:        aes.Key,
	Timeout:    time.Hour,
	MaxRefresh: time.Hour,
	HTTPStatusMessageFunc: func(err error, c *gin.Context) string {
		switch err {
		case jwt.ErrFailedAuthentication:
			return "incorrect Username, Password or Verification Code"
		default:
			return err.Error()
		}
	},
	Authenticator: func(email string, password string, c *gin.Context) (interface{}, bool) {
		twoFactorAuth := c.Query("verification_code")
		userFound, exists := services.UserExists(email, password)
		processLoginAttempt(email, exists, c)
		return &userFound, exists && services.TwoFactorMatchesAndIsNotOutdated(
			userFound.TwoFactorAuth.ValueOrZero(),
			twoFactorAuth,
			userFound.TwoFactorTime.ValueOrZero())
	},
	Authorizator: func(user interface{}, c *gin.Context) bool {
		userparam := c.Param("email")
		username := user.(string)
		return userparam == "" || userparam == username
	},
}

func GetUsers(c *gin.Context) {
	users, err := services.GetUsers()
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Unable to retrieve users: "+err.Error())
		return
	}
	var dtos []models.UserDto
	automapper.Map(users, &dtos)
	c.JSON(http.StatusOK, dtos)
}

func GetUser(c *gin.Context) {
	username := c.Param("email")
	user, err := services.GetUser(username)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+username)
		} else {
			sendJsonMsg(c, http.StatusInternalServerError, "Unable to find user: "+err.Error())
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
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		return
	}
	registeredUser, err := services.RegisterUser(&credentials)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to register user: "+err.Error())
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
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		return
	}
	username := c.Param("email")
	user, err = services.ModifyUser(username, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+username)
		} else {
			sendJsonMsg(c, http.StatusBadRequest, "Unable to modify user: "+err.Error())
		}
		return
	}
	var dto models.UserDto
	automapper.Map(user, &dto)
	c.JSON(http.StatusCreated, dto)
}

func DeleteUser(c *gin.Context) {
	email := c.Param("email")
	err := services.DeleteUser(email)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+email)
		} else {
			sendJsonMsg(c, http.StatusBadRequest, "Unable to delete user: "+err.Error())
		}
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func Generate2FAAndSendMail(c *gin.Context) {
	var credentials models.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		return
	}
	foundUser, exists := services.UserExists(credentials.Email, credentials.Password)
	processLoginAttempt(credentials.Email, exists, c)
	if !exists {
		sendJsonMsg(c, http.StatusNotFound, "Wrong credentials for: "+credentials.Email+". Please check the username and password are correct!")
		return
	}
	userWithCode, err := services.Generate2FACodeAndSaveToUser(&foundUser)
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Error creating secure code: "+err.Error())
		return
	}
	err = services.Send2FAEmail(userWithCode.Email, userWithCode.TwoFactorAuth.ValueOrZero())
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Error sending email: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, models.ServerError{
		Message: "Verification code sent correctly to " + credentials.Email + ". Check your inbox!",
	})
}

func processLoginAttempt(email string, successful bool, c *gin.Context) error {
	var attempt models.LoginAttempt
	attempt.User.Email = email
	attempt.UserAgent.SetValid(c.GetHeader(headers.UserAgent))
	attempt.IP.SetValid(c.ClientIP())
	attempt.Successful = successful
	return services.ProcessLoginAttempt(&attempt)
}

func sendJsonMsg(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, models.ServerError{
		Message: msg,
	})
}
