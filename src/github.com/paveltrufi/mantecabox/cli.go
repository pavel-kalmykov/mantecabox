package main

import (
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/go-http-utils/headers"
	"github.com/hako/durafmt"
	"github.com/howeyc/gopass"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/services"
	"gopkg.in/resty.v1"
)

func init() {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resty.SetHostURL("https://localhost:10443")
	resty.SetHeader(headers.ContentType, "application/json")
	resty.SetHeader(headers.Accept, "application/json")
}

func signup(credentialsFunc func() models.Credentials) error {
	fmt.Println("Welcome to mantecabox!")
	credentials := credentialsFunc()

	strength := zxcvbn.PasswordStrength(credentials.Password, []string{credentials.Username}).Score
	fmt.Printf("Password's strength: %v (out of 4).\n", strength)
	if strength <= 2 {
		return errors.New("password too guessable")
	}

	credentials.Password = hashAndEncodePassword(credentials.Password)

	err := services.ValidateCredentials(&credentials)
	if err != nil {
		return err
	}

	var result models.UserDto
	var serverError models.ServerError
	response, err := resty.R().
		SetBody(&credentials).
		SetResult(&result).
		SetError(&serverError).
		Post("/register")

	if err != nil {
		return err
	}
	if serverError.Message != "" {
		return errors.New(serverError.Message)
	}
	if response.StatusCode() != http.StatusCreated {
		return errors.New("server did not sent HTTP 201 Created status")
	}
	if result.Username != credentials.Username {
		return errors.New("username not registered properly")
	}

	fmt.Printf("User %v registered successfully!\n", result.Username)
	return nil
}

func login(credentialsFunc func() models.Credentials) error {
	fmt.Println("Nice to see you again!")
	credentials := credentialsFunc()
	credentials.Password = hashAndEncodePassword(credentials.Password)

	err := services.ValidateCredentials(&credentials)
	if err != nil {
		return err
	}

	var result models.JwtResponse
	var serverError models.ServerError
	response, err := resty.R().
		SetBody(&credentials).
		SetResult(&result).
		SetError(&serverError).
		Post("/login")

	if err != nil {
		return err
	}
	if serverError.Message != "" {
		return errors.New(serverError.Message)
	}
	if response.StatusCode() != http.StatusOK {
		return errors.New("server did not sent HTTP 200 OK status")
	}

	os.Setenv("MANTECABOX_TOKEN", result.Token)
	os.Setenv("MANTECABOX_TOKEN_EXPIRE", result.Expire.String())

	fmt.Printf("Successfully logged for %v", durafmt.ParseShort(result.Expire.Sub(time.Now())))
	return nil
}

func hashAndEncodePassword(password string) string {
	sum512 := sha512.Sum512([]byte(password))
	uppercasedHash := strings.ToUpper(hex.EncodeToString(sum512[:]))
	return base64.URLEncoding.EncodeToString([]byte(uppercasedHash))
}

func readCredentials() models.Credentials {
	var credentials models.Credentials
	fmt.Print("Username: ")
	fmt.Scanln(&credentials.Username)
	fmt.Print("Password: ")
	pass, err := gopass.GetPasswdMasked()
	if err != nil {
		panic(err)
	}
	credentials.Password = string(pass)
	return credentials
}

func main() {
	var args struct {
		Operation string `arg:"positional, required" help:"<signup>, <login>, <help>"`
	}
	parser := arg.MustParse(&args)
	if args.Operation == "signup" {
		err := signup(readCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during signup: %v.\n", err)
		}
	} else if args.Operation == "login" {
		err := login(readCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during login: %v.\n", err)
		}
	} else if args.Operation == "help" {
		parser.WriteHelp(os.Stdin)
	} else {
		parser.Fail(fmt.Sprintf(`Operation "%v" not recognized`, args.Operation))
	}
}
