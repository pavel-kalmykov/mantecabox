package webservice

import (
	"database/sql"
	"net/http"
	"time"

	"mantecabox/logs"
	"mantecabox/models"
	"mantecabox/services"

	"github.com/PeteProgrammer/go-automapper"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
)

type (
	UserController interface {
		GetUsers(c *gin.Context)
		GetUser(c *gin.Context)
		RegisterUser(c *gin.Context)
		ModifyUser(c *gin.Context)
		DeleteUser(c *gin.Context)
		Generate2FAAndSendMail(c *gin.Context)
		AuthMiddleware() *jwt.GinJWTMiddleware
	}

	UserControllerImpl struct {
		userService         services.UserService
		loginAttemptService services.LoginAttemptService
		mailService         services.MailService
		authMiddleware      *jwt.GinJWTMiddleware
	}
)

func NewUserController(configuration *models.Configuration) UserController {
	userService := services.NewUserService(configuration)
	loginAttemptService := services.NewLoginAttemptService(configuration)
	mailService := services.NewMailService(configuration)
	if userService == nil || loginAttemptService == nil || mailService == nil {
		return nil
	}
	tokenTimeout, err := time.ParseDuration(configuration.TokenTimeout)
	if err != nil {
		logs.ControllerLog.Fatal("Unable to parse token's timeout: " + err.Error())
	}
	return UserControllerImpl{
		userService:         userService,
		loginAttemptService: loginAttemptService,
		mailService:         mailService,
		authMiddleware: &jwt.GinJWTMiddleware{
			Realm:      "Mantecabox",
			Key:        userService.AesCipher().Key(),
			Timeout:    tokenTimeout,
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
				userFound, exists := userService.UserExists(email, password)
				var attempt models.LoginAttempt
				attempt.User.Email = email
				attempt.UserAgent.SetValid(c.GetHeader(headers.UserAgent))
				attempt.IP.SetValid(c.ClientIP())
				attempt.Successful = exists
				loginAttemptService.ProcessLoginAttempt(&attempt)
				return &userFound, exists && userService.TwoFactorMatchesAndIsNotOutdated(
					userFound.TwoFactorAuth.ValueOrZero(),
					twoFactorAuth,
					userFound.TwoFactorTime.ValueOrZero())
			},
			Authorizator: func(user interface{}, c *gin.Context) bool {
				userparam := c.Param("email")
				username := user.(string)
				return userparam == "" || userparam == username
			},
		},
	}
}

func (userController UserControllerImpl) GetUsers(c *gin.Context) {
	users, err := userController.userService.GetUsers()
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Unable to retrieve users: "+err.Error())
		logs.ControllerLog.Error("Unable to retrieve users: "+err.Error())
		return
	}
	var dtos []models.UserDto
	automapper.Map(users, &dtos)
	c.JSON(http.StatusOK, dtos)
}

func (userController UserControllerImpl) GetUser(c *gin.Context) {
	username := c.Param("email")
	user, err := userController.userService.GetUser(username)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+username)
			logs.ControllerLog.Error("Unable to find user: "+username)
		} else {
			sendJsonMsg(c, http.StatusInternalServerError, "Unable to find user: "+err.Error())
			logs.ControllerLog.Error("Unable to find user: "+err.Error())
		}
		return
	}
	var dto models.UserDto
	automapper.Map(user, &dto)
	c.JSON(http.StatusOK, dto)
}

func (userController UserControllerImpl) RegisterUser(c *gin.Context) {
	var credentials models.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		logs.ControllerLog.Error("Unable to parse JSON: "+err.Error())
		return
	}
	registeredUser, err := userController.userService.RegisterUser(&credentials)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to register user: "+err.Error())
		logs.ControllerLog.Error("Unable to register user: "+err.Error())
		return
	}
	var dto models.UserDto
	automapper.Map(registeredUser, &dto)
	c.JSON(http.StatusCreated, dto)
}

func (userController UserControllerImpl) ModifyUser(c *gin.Context) {
	var user models.User
	err := c.ShouldBindJSON(&user)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		logs.ControllerLog.Error("Unable to parse JSON: "+err.Error())
		return
	}
	username := c.Param("email")
	user, err = userController.userService.ModifyUser(username, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+username)
			logs.ControllerLog.Error("Unable to find user: "+username)
		} else {
			sendJsonMsg(c, http.StatusBadRequest, "Unable to modify user: "+err.Error())
			logs.ControllerLog.Error("Unable to modify user: "+err.Error())
		}
		return
	}
	var dto models.UserDto
	automapper.Map(user, &dto)
	c.JSON(http.StatusCreated, dto)
}

func (userController UserControllerImpl) DeleteUser(c *gin.Context) {
	email := c.Param("email")
	err := userController.userService.DeleteUser(email)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(c, http.StatusNotFound, "Unable to find user: "+email)
			logs.ControllerLog.Error("Unable to find user: "+email)
		} else {
			sendJsonMsg(c, http.StatusBadRequest, "Unable to delete user: "+err.Error())
			logs.ControllerLog.Error("Unable to delete user: "+err.Error())
		}
		return
	}
	c.Writer.WriteHeader(http.StatusNoContent)
}

func (userController UserControllerImpl) Generate2FAAndSendMail(c *gin.Context) {
	var credentials models.Credentials
	err := c.ShouldBindJSON(&credentials)
	if err != nil {
		sendJsonMsg(c, http.StatusBadRequest, "Unable to parse JSON: "+err.Error())
		logs.ControllerLog.Error("Unable to parse JSON: "+err.Error())
		return
	}
	foundUser, exists := userController.userService.UserExists(credentials.Email, credentials.Password)
	var attempt models.LoginAttempt
	attempt.User.Email = credentials.Email
	attempt.UserAgent.SetValid(c.GetHeader(headers.UserAgent))
	attempt.IP.SetValid(c.ClientIP())
	attempt.Successful = exists
	userController.loginAttemptService.ProcessLoginAttempt(&attempt)
	if !exists {
		sendJsonMsg(c, http.StatusNotFound, "Wrong credentials for: "+credentials.Email+". Please check the username and password are correct!")
		return
	}
	userWithCode, err := userController.userService.Generate2FACodeAndSaveToUser(&foundUser)
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Error creating secure code: "+err.Error())
		logs.ControllerLog.Error("Error creating secure code: "+err.Error())
		return
	}
	err = userController.mailService.Send2FAEmail(userWithCode.Email, userWithCode.TwoFactorAuth.ValueOrZero())
	if err != nil {
		sendJsonMsg(c, http.StatusInternalServerError, "Error sending email: "+err.Error())
		logs.ControllerLog.Error("Error sending email: "+err.Error())
		return
	}
	c.JSON(http.StatusOK, models.ServerError{
		Message: "Verification code sent correctly to " + credentials.Email + ". Check your inbox!",
	})
}

func (userController UserControllerImpl) AuthMiddleware() *jwt.GinJWTMiddleware {
	return userController.authMiddleware
}

func sendJsonMsg(c *gin.Context, status int, msg string) {
	c.AbortWithStatusJSON(status, models.ServerError{
		Message: msg,
	})
}
