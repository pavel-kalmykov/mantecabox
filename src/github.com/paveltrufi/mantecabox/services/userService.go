package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"github.com/badoux/checkmail"
	"github.com/paveltrufi/mantecabox/config"
	"github.com/paveltrufi/mantecabox/dao/factory"
	"github.com/paveltrufi/mantecabox/dao/interfaces"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/utilities/aes"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

var (
	sha512Regex          *regexp.Regexp
	userDao              interfaces.UserDao
	configuration        models.Configuration
	InvalidEmailError    = errors.New("invalid email")
	InvalidPasswordError = errors.New("password input is not SHA-512 hashed")
	Generating2FAError   = errors.New("unable to generate a 2FA secure code")
)

func init() {
	conf := config.GetServerConf()
	configuration = conf
	dao := factory.UserDaoFactory(configuration.Engine)
	userDao = dao

	sha512Compile, err := regexp.Compile(`^[A-Fa-f0-9]{128}$`)
	if err != nil {
		panic(err)
	}
	sha512Regex = sha512Compile
}

func GetUsers() ([]models.User, error) {
	return userDao.GetAll()
}

func GetUser(username string) (models.User, error) {
	return userDao.GetByPk(username)
}

func RegisterUser(c *models.Credentials) (models.User, error) {
	var user models.User
	if err := ValidateCredentials(c); err != nil {
		return user, err
	}
	decodedPassword, err := base64.URLEncoding.DecodeString(c.Password)
	if err != nil {
		return user, err
	}
	bcryptedPassword, err := bcrypt.GenerateFromPassword(decodedPassword, bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}
	user.Credentials = models.Credentials{
		Email:    c.Email,
		Password: base64.URLEncoding.EncodeToString(aes.Encrypt(bcryptedPassword)),
	}
	return userDao.Create(&user)
}

func ModifyUser(username string, u *models.User) (models.User, error) {
	var user models.User
	if err := ValidateCredentials(&u.Credentials); err != nil {
		return user, err
	}
	decodedPassword, err := base64.URLEncoding.DecodeString(u.Password)
	if err != nil {
		return user, err
	}
	bcryptedPassword, err := bcrypt.GenerateFromPassword(decodedPassword, bcrypt.DefaultCost)
	if err != nil {
		return user, err
	}
	user = models.User{
		TimeStamp:  u.TimeStamp,
		SoftDelete: u.SoftDelete,
		Credentials: models.Credentials{
			Email:    u.Email,
			Password: base64.URLEncoding.EncodeToString(aes.Encrypt(bcryptedPassword)),
		},
	}
	return userDao.Update(username, &user)
}

func DeleteUser(username string) error {
	return userDao.Delete(username)
}

func UserExists(username, password string) (string, bool) {
	user, err := userDao.GetByPk(username)
	if err != nil {
		return username, false
	}
	decodedExpectedPassword, err := base64.URLEncoding.DecodeString(password)
	if err != nil {
		return username, false
	}
	decodedActualPassword, err := base64.URLEncoding.DecodeString(user.Password)
	if err != nil {
		return username, false
	}
	err = bcrypt.CompareHashAndPassword(aes.Decrypt(decodedActualPassword), decodedExpectedPassword)
	if err != nil {
		return username, false
	}
	return username, true
}

func ValidateCredentials(c *models.Credentials) error {
	decodedPassword, err := base64.URLEncoding.DecodeString(c.Password)
	if err != nil {
		return err
	}
	if err := checkmail.ValidateFormat(c.Email); err != nil {
		return errors.New(fmt.Sprintf("%v (%v)", InvalidEmailError, err))
	}
	if matches := sha512Regex.Match(decodedPassword); !matches {
		return InvalidPasswordError
	}
	return nil
}
