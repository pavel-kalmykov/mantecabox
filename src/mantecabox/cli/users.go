package cli

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"mantecabox/models"
	"mantecabox/services"

	"github.com/go-resty/resty"
	"github.com/hako/durafmt"
	"github.com/nbutton23/zxcvbn-go"

	"github.com/gin-gonic/gin"
)

func Signup(credentialsFunc func() models.Credentials) error {
	fmt.Println("Welcome to mantecabox!")
	credentials := credentialsFunc()

	strength := zxcvbn.PasswordStrength(credentials.Password, []string{credentials.Email}).Score
	fmt.Printf("Password's strength: %v (out of 4).\n", strength)
	if strength <= 2 {
		return errors.New("password too guessable")
	}

	credentials.Password = HashAndEncodePassword(credentials.Password)

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
	if result.Email != credentials.Email {
		return errors.New("username not registered properly")
	}

	fmt.Println(SuccesMessage("User %v registered successfully!\n", result.Email))
	return nil
}

func Login(credentialsFunc func() models.Credentials) error {
	fmt.Println("Nice to see you again!")
	credentials := credentialsFunc()
	credentials.Password = HashAndEncodePassword(credentials.Password)

	err := services.ValidateCredentials(&credentials)
	if err != nil {
		return err
	}

	var verificationResult models.ServerError
	var serverError models.ServerError
	s := GetSpinner()
	response, err := resty.R().
		SetBody(&credentials).
		SetResult(&verificationResult).
		SetError(&serverError).
		Post("/2fa-verification")
	s.Stop()

	if err != nil {
		return err
	}
	if serverError.Message != "" {
		return errors.New(serverError.Message)
	}
	if response.StatusCode() != http.StatusOK {
		return errors.New("server did not sent HTTP 200 OK status")
	}

	var twoFactorAuth string
	fmt.Println(verificationResult.Message)
	fmt.Print("Verification Code: M-")
	fmt.Scanln(&twoFactorAuth)

	var result models.JwtResponse
	response, err = resty.R().
		SetBody(gin.H{"username": credentials.Email, "password": credentials.Password}).
		SetQueryParam("verification_code", twoFactorAuth).
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

	_, err = SetToken(result)
	if err != nil {
		return err
	}

	fmt.Println(SuccesMessage("Successfully logged for %v", durafmt.ParseShort(result.Expire.Sub(time.Now())).String()))
	return nil
}
