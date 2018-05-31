package webservice

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/labstack/gommon/log"
	"github.com/paveltrufi/mantecabox/models"
	"github.com/paveltrufi/mantecabox/services"
	"github.com/paveltrufi/mantecabox/utilities/aes"
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

/*
Función encargada de la subida y cifrado de los ficheros.
 */
func UploadFile(context *gin.Context) {

	path := "./files/"

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
		uploatedFile, error := services.CreateFile(&models.File {
			Name:  header.Filename,
			Owner: models.User{
				Credentials: models.Credentials {
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
		if err := ioutil.WriteFile(path + strconv.FormatInt(uploatedFile.Id, 10), encrypted, 0755); err != nil {
			sendJsonMsg(context, http.StatusBadRequest, err.Error())
			return
		}

		/*
		Enviamos una respuesta positiva al cliente
		 */
		context.Writer.WriteHeader(http.StatusCreated)
	}
}
