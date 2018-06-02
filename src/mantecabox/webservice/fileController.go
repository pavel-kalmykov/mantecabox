package webservice

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"

	"mantecabox/models"
	"mantecabox/services"

	"github.com/appleboy/gin-jwt"
	"github.com/benashford/go-func"
	"github.com/gin-gonic/gin"
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

	context.JSON(http.StatusCreated, models.FileToDto(fileModel))
}

func DeleteFile(context *gin.Context) {
	fileID := context.Param("file")

	file, err := strconv.ParseInt(fileID, 10, 64)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}

	err = services.DeleteFile(file, fileID)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+fileID)
		} else {
			sendJsonMsg(context, http.StatusBadRequest, "Unable to delete file: "+err.Error())
		}
		return
	}

	context.Writer.WriteHeader(http.StatusNoContent)
}

func GetFile(context *gin.Context) {

	filename := context.Param("file")

	file, err := checkSameFileExist(filename, getUser(context))
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		}
	}

	fileDecrypt, err := services.GetDecryptedLocalFile(file)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
	}

	contentLength, contentType, reader, extraHeaders := services.GetFileStream(fileDecrypt, file)

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

	filesDto := funcs.Maps(files, models.FileToDto).([]models.FileDTO)
	context.JSON(http.StatusOK, filesDto)
}
