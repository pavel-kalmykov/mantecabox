package main

import (
	"crypto/sha512"
	"crypto/tls"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"mantecabox/models"
	"mantecabox/services"

	"github.com/alexflint/go-arg"
	"github.com/appleboy/gin-jwt"
	"github.com/briandowns/spinner"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
	"github.com/hako/durafmt"
	"github.com/howeyc/gopass"
	"github.com/mitchellh/go-homedir"
	"github.com/nbutton23/zxcvbn-go"
	"github.com/tidwall/gjson"
	"github.com/zalando/go-keyring"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/resty.v1"

)

const (
	keyringServiceName = "mantecabox"
	loginToken         = "login_token"
)

func init() {
	resty.SetTLSClientConfig(&tls.Config{InsecureSkipVerify: true})
	resty.SetHostURL("https://localhost:10443")
	resty.SetHeader(headers.ContentType, "application/json")
	resty.SetHeader(headers.Accept, "application/json")

	home, err := homedir.Dir()
	if err != nil {
		panic(err)
	}
	resty.SetOutputDirectory(home + "/Mantecabox")
}

func signup(credentialsFunc func() models.Credentials) error {
	fmt.Println("Welcome to mantecabox!")
	credentials := credentialsFunc()

	strength := zxcvbn.PasswordStrength(credentials.Password, []string{credentials.Email}).Score
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
	if result.Email != credentials.Email {
		return errors.New("username not registered properly")
	}

	fmt.Printf("User %v registered successfully!\n", result.Email)
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

	var verificationResult models.ServerError
	var serverError models.ServerError
	s := spinner.New(spinner.CharSets[14], 100*time.Millisecond)
	s.Start()
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

	bytes, err := json.Marshal(result)
	if err != nil {
		return err
	}
	err = keyring.Set(keyringServiceName, loginToken, string(bytes))
	if err != nil {
		return err
	}

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

func getToken() (string, error) {
	/*
	TODO:
	1- Obtener el string del keyring (json).
	2- Parsear el json a la struct.
	3- Comprobar caducidad del token. {
		- Si la fecha de caducidad es anterior a la actual: caducado. Return error.
		- Si el token no está caducdado: {
			- Si le faltan 15 minutos: refrescar el token.
			- Sino lanzar la petición.
		}
	}
	*/

	tokenStr, err := keyring.Get(keyringServiceName, loginToken)
	if err != nil {
		return tokenStr, err
	}

	parsedToken := models.JwtResponse{}

	err = json.Unmarshal([]byte(tokenStr), &parsedToken)
	if err != nil {
		return "", err
	}

	if parsedToken.Expire.Before(time.Now()) {
		return "", jwt.ErrExpiredToken
	}

	return parsedToken.Token, nil
}

func getFiles(token string) ([]gjson.Result, error) {
	response, err := resty.R().
		SetAuthToken(token).
		Get("/files")
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == http.StatusOK {
		list := gjson.Get(response.String(), "#.name").Array()
		return  list, nil
	} else {
		return nil, errors.New("server did not sent HTTP 200 OK status. " + response.String())
	}
}

func uploadFile(filePath string, token string) (string, error) {
	response, err := resty.R().
		SetFiles(map[string]string{
		"file": filePath,
	}).
		SetAuthToken(token).
		Post("/files")

	if err != nil {
		return "", err
	}

	fileName := gjson.Get(response.String(), "name")

	if response.StatusCode() != http.StatusCreated {
		return "", errors.New(fmt.Sprintf("an error was received while uploading the '%v' file.", fileName))
	}

	return fileName.Str, nil
}

func downloadFile(fileSelected string, token string) error {
	response, err := resty.R().
		SetAuthToken(token).
		SetOutput(fileSelected).
		Get("/files/" + fileSelected)
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return errors.New(fmt.Sprintf("an error was received while downloading the '%v' file.", fileSelected))
	}

	return nil
}

func deleteFile(filename string, token string) error {
	response, err := resty.R().
		SetAuthToken(token).
		Delete("/files/" + filename)
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusNoContent {
		return errors.New(fmt.Sprintf("an error was received while removing the '%v' file.", filename))
	}

	return nil
}

func transfer(transferActions [] string) error {

	token, err := getToken()
	if err != nil {
		return err
	}

	lengthActions := len(transferActions)

	if lengthActions > 0 {
		switch transferActions[0] {
		case "list":
			list, err := getFiles(token)
			if err != nil {
				return err
			}

			fmt.Println("These are your files:")
			for _, f := range list {
				fmt.Printf(" - %v\n", f)
			}
		case "upload":
			if lengthActions > 1 {
				for i := 1; i < len(transferActions); i++ {
					fileName, err := uploadFile(transferActions[i], token)
					if err != nil {
						return err
					}

					fmt.Printf("File '%v' has uploaded correctly.\n", fileName)
				}
			} else {
				return errors.New(fmt.Sprintf("params not found"))
			}
		case "download":
			if lengthActions > 1 {
				for i := 1; i < len(transferActions); i++ {
					err := downloadFile(transferActions[i], token)
					if err != nil {
						return err
					}

					fmt.Printf("File '%v' has downloaded correctly.\n", transferActions[i])
				}
			} else {
				list, err := getFiles(token)
				if err != nil {
					return err
				}

				var listaString []string
				for _, f := range list {
					listaString = append(listaString, f.Str)
				}

				fileSelected := ""
				prompt := &survey.Select{
					Message: "These are your files. Please, choose once: ",
					Options: listaString,
				}
				survey.AskOne(prompt, &fileSelected, nil)

				err = downloadFile(fileSelected, token)
				if err != nil {
					return err
				}
				fmt.Printf("File '%v' has downloaded correctly.", fileSelected)
			}
		case "remove":
			if lengthActions > 1 {
				for i := 1; i < len(transferActions); i++ {
					err := deleteFile(transferActions[i], token)
					if err != nil {
						return err
					}

					fmt.Printf("File '%v' has removed correctly.\n", transferActions[i])
				}
			} else {
				return errors.New(fmt.Sprintf("params not found"))
			}
		default:
			return errors.New(fmt.Sprintf("action '%v' not exist", transferActions[0]))
		}
	} else {
		return errors.New(fmt.Sprintf("action '%v' not found", transferActions[0]))
	}

	return nil
}

func main() {
	var args struct {
		Operation       string   `arg:"positional, required" help:"(signup|login|transfer|help)"`
		TransferActions []string `arg:"positional" help:"(list|((upload|download|remove) <files>...)"`
	}
	parser := arg.MustParse(&args)

	switch args.Operation {
	case "signup":
		err := signup(readCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during signup: %v\n", err)
		}
		break
	case "login":
		err := login(readCredentials)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during login: %v\n", err)
		}
		break
	case "transfer":
		err := transfer(args.TransferActions)
		if err != nil {
			fmt.Fprintf(os.Stderr, "An error ocurred during transfer: %v\n", err)
		}
		break
	case "help":
		parser.WriteHelp(os.Stdin)
		break
	default:
		parser.Fail(fmt.Sprintf(`Operation "%v" not recognized`, args.Operation))
		break
	}
}
