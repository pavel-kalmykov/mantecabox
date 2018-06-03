package cli

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/tidwall/gjson"
	"gopkg.in/AlecAivazis/survey.v1"
	"gopkg.in/resty.v1"
)

func uploadFile(filePath string, token string) (string, error) {

	s := GetSpinner()
	response, err := resty.R().
		SetFiles(map[string]string{
		"file": filePath,
	}).
		SetAuthToken(token).
		Post("/files")
	s.Stop()

	if err != nil {
		return "", err
	}

	fileName := gjson.Get(response.String(), "name")

	if response.StatusCode() != http.StatusCreated && response.StatusCode() != http.StatusOK {
		return "", errors.New(fmt.Sprintf("an error was received while uploading the '%v' file.", fileName))
	}

	return fileName.Str, nil
}

func downloadFile(fileSelected string, token string) error {
	s := GetSpinner()
	response, err := resty.R().
		SetAuthToken(token).
		SetOutput(fileSelected).
		Get("/files/" + fileSelected)
	s.Stop()
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusOK {
		return errors.New(fmt.Sprintf("an error was received while downloading the '%v' file.", fileSelected))
	}

	return nil
}

func deleteFile(filename string, token string) error {
	s := GetSpinner()
	response, err := resty.R().
		SetAuthToken(token).
		Delete("/files/" + filename)
	s.Stop()
	if err != nil {
		return err
	}

	if response.StatusCode() != http.StatusNoContent {
		return errors.New(fmt.Sprintf("an error was received while removing the '%v' file.", filename))
	}

	return nil
}

func Transfer(transferActions [] string) error {
	token, err := GetToken()
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
						fmt.Printf("Error has succes in intent of upload file '%v' \n", transferActions[i])
					} else {
						fmt.Printf("File '%v' has uploaded correctly.\n", fileName)
					}
				}
			} else {
				return errors.New(fmt.Sprintf("params not found"))
			}
		case "download":
			if lengthActions > 1 {
				for i := 1; i < len(transferActions); i++ {
					err := downloadFile(transferActions[i], token)
					if err != nil {
						fmt.Printf("Error has succes in intent of download file '%v' \n", transferActions[i])
					} else {
						fmt.Printf("File '%v' has downloaded correctly.\n", transferActions[i])
					}
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

func getFiles(token string) ([]gjson.Result, error) {
	s := GetSpinner()
	response, err := resty.R().
		SetAuthToken(token).
		Get("/files")
	s.Stop()
	if err != nil {
		return nil, err
	}
	if response.StatusCode() == http.StatusOK {
		list := gjson.Get(response.String(), "#.name").Array()
		return list, nil
	} else {
		return nil, errors.New("server did not sent HTTP 200 OK status. " + response.String())
	}
}
