package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"mantecabox/dao"
	"mantecabox/models"
	"mantecabox/utilities"

	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
)

var (
	sha512Regex          *regexp.Regexp
	InvalidEmailError    = errors.New("invalid email")
	InvalidPasswordError = errors.New("password input is not SHA-512 hashed")
	Generating2FAError   = errors.New("unable to generate a 2FA secure code")
	Empty2FACodeError    = errors.New("the 2FA secure code is empty")
)

type (
	UserService interface {
		GetUsers() ([]models.User, error)
		GetUser(username string) (models.User, error)
		RegisterUser(c *models.Credentials) (models.User, error)
		ModifyUser(username string, u *models.User) (models.User, error)
		DeleteUser(username string) error
		UserExists(email, password string) (models.User, bool)
		Generate2FACodeAndSaveToUser(user *models.User) (models.User, error)
		TwoFactorMatchesAndIsNotOutdated(expected, actual string, expire time.Time) bool
		UserDao() dao.UserDao
		AesCipher() utilities.AesCTRCipher
	}

	UserServiceImpl struct {
		configuration *models.Configuration
		userDao       dao.UserDao
		aesCipher     utilities.AesCTRCipher
	}
)

func init() {
	sha512Compile, err := regexp.Compile(`^[A-Fa-f0-9]{128}$`)
	if err != nil {
		panic(err)
	}
	sha512Regex = sha512Compile
}

func NewUserService(configuration *models.Configuration) UserService {
	if configuration == nil {
		return nil
	}
	return UserServiceImpl{
		configuration: configuration,
		userDao:       dao.UserDaoFactory(configuration.Database.Engine),
		aesCipher:     utilities.NewAesCTRCipher(configuration.AesKey),
	}
}

func (userService UserServiceImpl) GetUsers() ([]models.User, error) {
	return userService.userDao.GetAll()
}

func (userService UserServiceImpl) GetUser(username string) (models.User, error) {
	return userService.userDao.GetByPk(username)
}

func (userService UserServiceImpl) RegisterUser(c *models.Credentials) (models.User, error) {
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
		Password: base64.URLEncoding.EncodeToString(userService.aesCipher.Encrypt(bcryptedPassword)),
	}
	return userService.userDao.Create(&user)
}

func (userService UserServiceImpl) ModifyUser(username string, u *models.User) (models.User, error) {
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
			Password: base64.URLEncoding.EncodeToString(userService.aesCipher.Encrypt(bcryptedPassword)),
		},
	}
	return userService.userDao.Update(username, &user)
}

func (userService UserServiceImpl) DeleteUser(username string) error {
	return userService.userDao.Delete(username)
}

func (userService UserServiceImpl) UserExists(email, password string) (models.User, bool) {
	user, err := userService.userDao.GetByPk(email)
	if err != nil {
		user.Email = email
		return user, false
	}
	decodedExpectedPassword, err := base64.URLEncoding.DecodeString(password)
	if err != nil {
		return user, false
	}
	decodedActualPassword, err := base64.URLEncoding.DecodeString(user.Password)
	if err != nil {
		return user, false
	}
	err = bcrypt.CompareHashAndPassword(userService.aesCipher.Decrypt(decodedActualPassword), decodedExpectedPassword)
	if err != nil {
		return user, false
	}
	return user, true
}

func (userService UserServiceImpl) Generate2FACodeAndSaveToUser(user *models.User) (models.User, error) {
	secureCode, err := rand.Int(rand.Reader, big.NewInt(999999)) // 6 digits max
	if err != nil {
		return *user, Generating2FAError
	}
	paddedSecureCode := fmt.Sprintf("%06d", secureCode)
	user.TwoFactorAuth.SetValid(paddedSecureCode)
	return userService.userDao.Update(user.Email, user)
}

func (userService UserServiceImpl) TwoFactorMatchesAndIsNotOutdated(expected, actual string, expire time.Time) bool {
	duration, err := time.ParseDuration(userService.configuration.VerificationMailTimeLimit)
	if err != nil {
		panic("unable to parse mail's verification limit configuration value: " + err.Error())
	}
	timeLimit := duration
	return expected == actual && time.Now().Sub(expire) < timeLimit
}

func (userService UserServiceImpl) UserDao() dao.UserDao {
	return userService.userDao
}

func (userService UserServiceImpl) AesCipher() utilities.AesCTRCipher {
	return userService.aesCipher
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
