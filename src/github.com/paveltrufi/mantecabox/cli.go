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

	"github.com/alexflint/go-arg"
	"github.com/go-http-utils/headers"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/paveltrufi/mantecabox/models"
	"gopkg.in/resty.v1"
)

func init() {
	resty.SetHostURL("https://localhost:10443")
	resty.SetHeader(headers.ContentType, "application/json")
	resty.SetHeader(headers.Accept, "application/json")
}

func signup() error {
	var credentials models.Credentials

	fmt.Println("Welcome to Mantecabox")
	fmt.Print("Username: ")
	fmt.Scanln(&credentials.Username)
	fmt.Print("Password: ")
	fmt.Scanln(&credentials.Password)

	strength := zxcvbn.PasswordStrength(credentials.Password, []string{credentials.Username}).Score
	fmt.Printf("Password's strength: %v (out of 4).\n", strength)
	if strength <= 2 {
		return errors.New("password too guessable")
	}

	sum512 := sha512.Sum512([]byte(credentials.Password))
	str := strings.ToUpper(hex.EncodeToString(sum512[:]))
	credentials.Password = base64.URLEncoding.EncodeToString([]byte(str))

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

func main() {
	resty.SetTLSClientConfig(&tls.Config{ InsecureSkipVerify: true })
	var args struct {
		Operation string `arg:"positional, required" help:"<signup>, <login>"`
	}
	arg.MustParse(&args)
	if args.Operation == "signup" {
		err := signup()
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during signup: %v.\n", err)
		}
	} else if args.Operation == "login" {

	}
}
