package webservice

import (
	"bytes"
	"database/sql"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"mantecabox/models"
	"mantecabox/services"
	"mantecabox/utilities/aes"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
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

var path = "./files/"

/*
Función encargada de la subida y cifrado de los ficheros.
 */
func UploadFile(context *gin.Context) {
	/*
	Obtención del fichero desde el post
	 */
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		sendJsonMsg(context, http.StatusBadRequest, err.Error())
		return
	}

	/*
	Comprobación o creación del directorio files
	 */
	if CreateDirIfNotExist(path) {
		/*
		- Obtención del username desde el token
		- Creación del modelo de fichero que vamos a enviar a la base de datos
		- Creación del fichero en la base de datos
		 */
		uploatedFile, error := services.CreateFile(&models.File{
			Name: header.Filename,
			Owner: models.User{
				Credentials: models.Credentials{
					Email: jwt.ExtractClaims(context)["id"].(string),
				}},
		})
		if error != nil {
			sendJsonMsg(context, http.StatusBadRequest, error.Error())
			return
		}

		/*
		Conversión a bytes del fichero
		 */
		buf := bytes.NewBuffer(nil)
		if _, err := io.Copy(buf, file); err != nil {
			sendJsonMsg(context, http.StatusBadRequest, err.Error())
			return
		}

		/*
		Encriptado del fichero
		 */
		encrypted := aes.Encrypt(buf.Bytes())

		/*
		Guardamos el fichero encriptado
		 */
		if err := ioutil.WriteFile(path+strconv.FormatInt(uploatedFile.Id, 10), encrypted, 0755); err != nil {
			sendJsonMsg(context, http.StatusBadRequest, err.Error())
			return
		}

		/*
		Enviamos una respuesta positiva al cliente
		 */
		context.Writer.WriteHeader(http.StatusCreated)
	}
}

func DeleteFile(context *gin.Context) {
	fileID := context.Param("file")

	file, error := strconv.ParseInt(fileID, 10, 64)
	if error != nil {
		sendJsonMsg(context, http.StatusInternalServerError, error.Error())
		return
	}

	err := services.DeleteFile(file)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+fileID)
		} else {
			sendJsonMsg(context, http.StatusBadRequest, "Unable to delete file: "+err.Error())
		}
		return
	}

	error = os.Remove(path + fileID)
	if error != nil {
		sendJsonMsg(context, http.StatusInternalServerError, error.Error())
		return
	}

	context.Writer.WriteHeader(http.StatusNoContent)
}

func GetFile(context *gin.Context) {

	filename := context.Param("file")

	user := models.User{
		Credentials: models.Credentials{
			Email: jwt.ExtractClaims(context)["id"].(string),
		}}

	file, err := services.DownloadFile(filename, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+filename)
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, "Unable to find file: "+err.Error())
		}

		return
	}

	fileEncrypt, error := ioutil.ReadFile(path + strconv.FormatInt(file.Id, 10))
	if error != nil {
		sendJsonMsg(context, http.StatusInternalServerError, error.Error())
		return
	}

	fileDecrypt := aes.Decrypt(fileEncrypt)

	reader := bytes.NewReader(fileDecrypt)
	contentLength := reader.Size()
	contentType := "application/octet-stream"

	extraHeaders := map[string]string{
		headers.ContentDisposition: `attachment; filename="` + file.Name + `"`,
	}

	context.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}
