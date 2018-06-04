package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"mantecabox/models"

	"github.com/mitchellh/go-homedir"
	"github.com/phayes/permbits"
	"github.com/radovskyb/watcher"
	"gopkg.in/guregu/null.v3"
)

var mantecaboxDir = ""

func monitorChanges(w *watcher.Watcher) {
	for {
		select {
		case event := <-w.Event:
			fmt.Println(event) // Print the event's info.
			fileInfos, err := ioutil.ReadDir(event.Path)
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
			}

			localFiles := make([]models.FileDTO, 0)
			for _, fileInfo := range fileInfos {
				if fileInfo.Name() != "" {
					localFiles = append(localFiles, models.FileDTO{
						TimeStamp:      models.TimeStamp{UpdatedAt: null.Time{Time: fileInfo.ModTime(), Valid: true}},
						Name:           fileInfo.Name(),
						PermissionsStr: permbits.FileMode(fileInfo.Mode()).String(),
					})
				}
			}

			token, err := GetToken()
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to get token:", err)
				return
			}

			ids, names, dates, permissions, err := getFilesList(token)
			if err != nil && err.Error() != "there are no files in our servers. Upload one" {
				fmt.Fprintln(os.Stderr, "unable to retrieve files:", err)
				return
			}
			remoteFiles := make([]models.FileDTO, 0)
			for i := 0; i < len(ids); i++ {
				remoteFiles = append(remoteFiles, models.FileDTO{
					TimeStamp:      models.TimeStamp{UpdatedAt: null.Time{Time: dates[i].Time(), Valid: true}},
					Name:           names[i].String(),
					PermissionsStr: permissions[i].String(),
				})
			}

			compareFileLists(localFiles, remoteFiles, token)
		case err := <-w.Error:
			fmt.Fprintln(os.Stderr, err)
		case <-w.Closed:
			return
		}
	}
}

func compareFileLists(localFiles []models.FileDTO, remoteFiles []models.FileDTO, token string) {
	for _, localFile := range localFiles {
		if !fileInSlice(localFile, remoteFiles) {
			filePath := mantecaboxDir + localFile.Name
			_, err := uploadFile(filePath, token)
			if err != nil {
				fmt.Fprintln(os.Stderr, "unable to upload file:", err)
				return
			}
			fmt.Println("File", localFile.Name, "uploaded")
		}
	}
}

func fileInSlice(a models.FileDTO, list []models.FileDTO) bool {
	for _, b := range list {
		if b == a {
			return true
		}
	}
	return false
}

func StartDaemon() {
	w := watcher.New()
	w.SetMaxEvents(1)
	go monitorChanges(w)
	dir, err := homedir.Dir()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	dir += "/Mantecabox/"
	mantecaboxDir = dir
	if err := w.Add(mantecaboxDir); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}

	fmt.Printf("Daemon started. Syncing in %v...\n", dir)
	if err := w.Start(time.Millisecond * 1000); err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
