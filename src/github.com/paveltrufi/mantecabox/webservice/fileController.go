package webservice

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"fmt"
	"os"
	"github.com/labstack/gommon/log"
)
func CreateDirIfNotExist(dir string) bool {
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		err = os.MkdirAll(dir, 0755)
		if err != nil {
			log.Print("Error en la creación del directorio")
			panic(err)
			return false
		}
	}

	return true
}

func UploadFile(c *gin.Context) {

	file, err := c.FormFile("file")

	if err != nil {
		c.String(http.StatusBadRequest, fmt.Sprintf("get form err: %s", err.Error()))
		return
	}

	// Data stock TODO: remove when connecting with the user session
	username := "user"
	path := "./files/" + username + "/" + file.Filename

	if CreateDirIfNotExist(path) {
		if err := c.SaveUploadedFile(file, path); err != nil {
			c.String(http.StatusBadRequest, fmt.Sprintf("Upload file err: %s", err.Error()))
			return
		}

		c.String(http.StatusOK, fmt.Sprintf("File %s uploaded successfully with fields.", file.Filename))
	}
}
