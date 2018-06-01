package webservice

import (
	"bytes"
	"database/sql"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"mantecabox/models"
	"mantecabox/services"
	"mantecabox/utilities/aes"

	"github.com/PeteProgrammer/go-automapper"
	"github.com/appleboy/gin-jwt"
	"github.com/gin-gonic/gin"
	"github.com/go-http-utils/headers"
)

// UploadFile se encarga de la subida y cifrado de los ficheros.
func UploadFile(context *gin.Context) {
	// Obtención del fichero desde el post
	file, header, err := context.Request.FormFile("file")
	if err != nil {
		sendJsonMsg(context, http.StatusBadRequest, err.Error())
		return
	}

	fileModel, err := checkSameFileExist(header.Filename, getUser(context))
	if err != nil {
		if err != sql.ErrNoRows {
			sendJsonMsg(context, http.StatusInternalServerError, "Unable to upload file: "+err.Error())
			return
		}
		fileModel, err = services.CreateFile(&models.File{
			Name:  header.Filename,
			Owner: getUser(context),
		})
		if err != nil {
			sendJsonMsg(context, http.StatusInternalServerError, err.Error())
			return
		}
	} else {
		fileModel, err = services.UpdateFile(fileModel.Id, fileModel)
		if err != nil {
			sendJsonMsg(context, http.StatusInternalServerError, err.Error())
			return
		}
	}
	err = services.SaveFile(file, fileModel)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
	}
	context.JSON(http.StatusCreated, models.FileDTO{
		Name:  fileModel.Name,
		Owner: fileModel.Owner.Email,
	})
}

// TODO mover cosas de negocio a la capa services
func DeleteFile(context *gin.Context) {
	fileID := context.Param("file")

	file, err := strconv.ParseInt(fileID, 10, 64)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}

	err = services.DeleteFile(file)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+fileID)
		} else {
			sendJsonMsg(context, http.StatusBadRequest, "Unable to delete file: "+err.Error())
		}
		return
	}

	err = os.Remove(services.Path + fileID)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}

	context.Writer.WriteHeader(http.StatusNoContent)
}

// TODO mover cosas de negocio a la capa services
func GetFile(context *gin.Context) {

	filename := context.Param("file")

	file, err := checkSameFileExist(filename, getUser(context))
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+filename)
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, "Unable to find file: "+err.Error())
		}
		return
	}

	fileEncrypt, err := ioutil.ReadFile(services.Path + strconv.FormatInt(file.Id, 10))
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
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

func getUser(context *gin.Context) models.User {
	var user models.User
	user.Email = jwt.ExtractClaims(context)["id"].(string)
	return user
}

func checkSameFileExist(filename string, user models.User) (file models.File, err error) {
	return services.GetFile(filename, &user)
}

func GetAllFiles(context *gin.Context) {
	files, err := services.GetAllFiles(getUser(context))
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to retrieve files: "+err.Error())
		return
	}

	var dtos []models.Files
	automapper.Map(files, &dtos)
	context.JSON(http.StatusOK, dtos)
}
