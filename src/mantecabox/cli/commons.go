package cli

import (
	"crypto/sha512"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"mantecabox/models"

	"github.com/appleboy/gin-jwt"
	"github.com/briandowns/spinner"
	"github.com/fatih/color"
	"github.com/go-resty/resty"
	"github.com/howeyc/gopass"
	"github.com/tidwall/gjson"
	"github.com/zalando/go-keyring"
)

const (
	keyringServiceName = "mantecabox"
	loginToken         = "login_token"
)

func SetToken(response interface{}) (string, error) {
	bytes, err := json.Marshal(response)
	if err != nil {
		return "", err
	}

	err = keyring.Set(keyringServiceName, loginToken, string(bytes))
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

func GetToken() (string, error) {
	tokenStr, err := keyring.Get(keyringServiceName, loginToken)
	if err != nil {
		return tokenStr, err
	}

	var parsedToken models.JwtResponse
	err = json.Unmarshal([]byte(tokenStr), &parsedToken)
	if err != nil {
		return "", err
	}

	if remainingTime := parsedToken.Expire.Sub(time.Now()); remainingTime < (time.Minute * 15) {
		var result models.JwtResponse
		response, err := resty.R().
			SetAuthToken(parsedToken.Token).
			SetResult(&result).
			Get("/refresh-token")

		if err != nil {
			return "", err
		}

		if response.StatusCode() == http.StatusUnauthorized {
			return "", jwt.ErrExpiredToken
		}

		token, err := SetToken(result)
		if err != nil {
			return "", err
		}

		tokenValue := gjson.Get(token, "token")
		return tokenValue.Str, nil
	}

	if parsedToken.Expire.Before(time.Now()) {
		return "", nil
	}

	return parsedToken.Token, nil
}

func GetSpinner() *spinner.Spinner {
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()
	return s
}

func HashAndEncodePassword(password string) string {
	sum512 := sha512.Sum512([]byte(password))
	uppercasedHash := strings.ToUpper(hex.EncodeToString(sum512[:]))
	return base64.URLEncoding.EncodeToString([]byte(uppercasedHash))
}

func ReadCredentials() models.Credentials {
	var credentials models.Credentials
	fmt.Print("Email: ")
	fmt.Scanln(&credentials.Email)
	fmt.Print("Password: ")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}
	credentials.Password = string(pass)
	return credentials
}

func SuccesMessage(text string, args ...string) string {
	c := color.New(color.FgCyan)
	return c.Sprintf("✔ "+text, strings.Join(args, ""))
}

func ErrorMessage(text string, args ...string) string {
	c := color.New(color.FgRed)
	return c.Sprintf("❌ "+text, strings.Join(args, ""))
}
