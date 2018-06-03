package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

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
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, authCode)
	if err != nil {
		log.Fatalf("Unable to retrieve token from web %v", err)
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
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	json.NewEncoder(f).Encode(token)
}

func main() {
	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved client_secret.json.
	config, err := google.ConfigFromJSON(b, drive.DriveMetadataReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}

	srv, err := drive.New(getClient(config))
	if err != nil {
		log.Fatalf("Unable to retrieve Drive client: %v", err)

	}

	// filedId := "1c2kXskYezmCSAdVev0SJnGDibK6JTXVJ" // get one available in file list
	// filename := "configuration.json"
	// filename := "configuration.test.json"
	// file, err := os.Open(filename)
	// if err != nil {
	// 	panic(err)
	// }

	// File download
	// driveFile, err := srv.Files.Get(filedId).Do()
	// if err != nil {
	// 	log.Fatalf("Unable to get file: %v", err)
	// }
	// log.Printf("got file: %+v (%v)", driveFile.Name, driveFile.Id)
	// response, err := srv.Files.Get(filedId).Download()
	// if err != nil {
	// 	log.Fatalf("Unable to download file: %v", err)
	// }
	// outFile, err := os.Create("downloaded.json")
	// if err != nil {
	// 	log.Fatalf("Unable to create file: %v", err)
	// }
	// defer outFile.Close()
	// _, err = io.Copy(outFile, response.Body)
	// if err != nil {
	// 	log.Fatalf("Unable to write file: %v", err)
	// }
	// log.Println("done")

	// File upload
	// driveFile, err := srv.Files.Create(&drive.File{Name: filename}).Media(file).Do()
	// if err != nil {
	// 	log.Fatalf("Unable to create file: %v", err)
	// }
	// log.Printf("uploaded file: %+v (%v)", driveFile.Name, driveFile.Id)
	// log.Println("done")

	// File update
	// driveFile, err := srv.Files.Update(filedId, &drive.File{Name: filename}).Media(file).Do()
	// if err != nil {
	// 	log.Fatalf("Unable to update file: %v", err)
	// }
	// log.Printf("updated file: %+v (%v)", driveFile.Name, driveFile.Id)
	// log.Println("done")

	// File delete
	// err = srv.Files.Delete(filedId).Do()
	// if err != nil {
	// 	log.Fatalf("Unable to delete file: %v", err)
	// }

	// File list
	r, err := srv.Files.List().
		Fields("nextPageToken, files(id, name)").Do()
	if err != nil {
		log.Fatalf("Unable to retrieve files: %v", err)
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
