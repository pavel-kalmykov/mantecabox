package services

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"math/big"
	"regexp"
	"time"

	"mantecabox/dao/postgres"
	"mantecabox/models"
	"mantecabox/utilities"
	"mantecabox/utilities/aes"

	"github.com/badoux/checkmail"
	"golang.org/x/crypto/bcrypt"
	"gopkg.in/gomail.v2"
)

var (
	sha512Regex          *regexp.Regexp
	configuration        = models.Configuration{}
	userDao              = postgres.UserPgDao{}
	InvalidEmailError    = errors.New("invalid email")
	InvalidPasswordError = errors.New("password input is not SHA-512 hashed")
	Generating2FAError   = errors.New("unable to generate a 2FA secure code")
	Empty2FACodeError    = errors.New("the 2FA secure code is empty")
)

func init() {
	sha512Compile, err := regexp.Compile(`^[A-Fa-f0-9]{128}$`)
	if err != nil {
		panic(err)
	}
	sha512Regex = sha512Compile
	configuration, err = utilities.GetConfiguration()
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

func UserExists(email, password string) (models.User, bool) {
	user, err := userDao.GetByPk(email)
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
	err = bcrypt.CompareHashAndPassword(aes.Decrypt(decodedActualPassword), decodedExpectedPassword)
	if err != nil {
		return user, false
	}
	return user, true
}

func Generate2FACodeAndSaveToUser(user *models.User) (models.User, error) {
	secureCode, err := rand.Int(rand.Reader, big.NewInt(999999)) // 6 digits max
	if err != nil {
		return *user, Generating2FAError
	}
	paddedSecureCode := fmt.Sprintf("%06d", secureCode)
	user.TwoFactorAuth.SetValid(paddedSecureCode)
	return userDao.Update(user.Email, user)
}

func Send2FAEmail(toEmail, code string) error {
	if code == "" {
		return Empty2FACodeError
	}
	if err := checkmail.ValidateHost(toEmail); err != nil {
		return err
	}
	return SendMail(toEmail, fmt.Sprintf("Hello. Your security code is M-<b>%v</b>. It will expire in 5 minutes", code))
}

func SendMail(toEmail, bodyMessage string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "mantecabox@gmail.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Mantecabox Backup")
	m.SetBody("text/html", bodyMessage)
	return gomail.
		NewDialer("smtp.gmail.com", 587, "mantecabox@gmail.com", "ElPutoPavel").
		DialAndSend(m)
}

func TwoFactorMatchesAndIsNotOutdated(expected, actual string, expire time.Time) bool {
	return expected == actual && time.Now().Sub(expire) < time.Minute*5
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
