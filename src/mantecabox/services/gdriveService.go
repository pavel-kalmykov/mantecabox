package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"

	"mantecabox/models"

	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
)

// Retrieve a token, saves the token, then returns the generated client.
func getClient(config *oauth2.Config) *http.Client {
	tokenFile := "token.json"
	tok, err := tokenFromFile(tokenFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(tokenFile, tok)
	}
	return config.Client(context.Background(), tok)
}

// Request a token from the web, then returns the retrieved token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var authCode string
	if _, err := fmt.Scan(&authCode); err != nil {
		logrus.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(context.Background(), authCode)
	if err != nil {
		logrus.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// Retrieves a token from a local file.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	defer f.Close()
	if err != nil {
		return nil, err
	}
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

// Saves a token to a file path.
func saveToken(path string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	defer f.Close()
	if err != nil {
		logrus.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func GetGdriveService() (*drive.Service, error) {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to read client secret file: %v", err))
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to parse client secret file to config: %v", err))
	}

	srv, err := drive.New(getClient(config))
	if err != nil {
		return nil, errors.New(fmt.Sprintf("Unable to retrieve Drive client: %v", err))
	}

	return srv, err
}

// File list
func ListFiles(srv *drive.Service) {
	r, err := srv.Files.List().
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		logrus.Fatalf("Unable to retrieve files: %v", err)
	}
	fmt.Println("Files:")
	if len(r.Files) == 0 {
		fmt.Println("No files found.")
	} else {
		for _, i := range r.Files {
			fmt.Printf("%s (%s)\n", i.Name, i.Id)
		}
	}
}

// File delere
func RemoveFile(srv *drive.Service, filedId string) error {
	err := srv.Files.Delete(filedId).Do()
	if err != nil {
		return err
	}

	return nil
}

// File upload
func (fileService FileServiceImpl) UploadFileGDrive(srv *drive.Service, filename string, file multipart.File) (*drive.File, error) {

	fileEncrypted, err := fileService.encryptFile(file)
	if err != nil {
		return nil, err
	}

	driveFile, err := srv.Files.Create(&drive.File{Name: filename}).Media(bytes.NewReader(fileEncrypted)).Do()
	if err != nil {
		logrus.Fatalf("Unable to create file: %v", err)
		return nil, err
	}

	logrus.Printf("uploaded file: %+v (%v)", driveFile.Name, driveFile.Id)

	return driveFile, err
}

// File update
func UpdateFile(srv *drive.Service, filedId string, filename string, file io.Reader) error {

	_, err := srv.Files.Update(filedId, &drive.File{Name: filename}).Media(file).Do()
	if err != nil {
		return err
	}

	return nil
}

// File download
func (fileService FileServiceImpl) DownloadFile(srv *drive.Service, filedId string, file models.File) ([]byte, error) {
	_, err := srv.Files.Get(filedId).Do()
	if err != nil {
		return nil, err
	}

	response, err := srv.Files.Get(filedId).Download()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	fileDecrypt := fileService.aesCipher.Decrypt(body)

	return fileDecrypt, nil
}

func (fileService FileServiceImpl) GetFileGDrive(file models.File) ([]byte, error) {
	gService, err := GetGdriveService()
	if err != nil {
		return nil, err
	}

	fileResponse, err := fileService.DownloadFile(gService, file.GdriveID.String, file)
	if err != nil {
		return nil, err
	}

	return fileResponse, err
}
