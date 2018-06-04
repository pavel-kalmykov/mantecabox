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

type (
	FileController interface {
		GetAllFiles(context *gin.Context)
		GetAllFileVersions(context *gin.Context)
		GetFile(context *gin.Context)
		GetFileVersion(context *gin.Context)
		DownloadFile(context *gin.Context)
		DownloadFileVersion(context *gin.Context)
		download(filename string, file models.File, err error, context *gin.Context)
		UploadFile(context *gin.Context)
		DeleteFile(context *gin.Context)
	}

	FileControllerImpl struct {
		configuration *models.Configuration
		fileService   services.FileService
	}
)

func NewFileController(configuration *models.Configuration) FileController {
	fileService := services.NewFileService(configuration)
	if fileService == nil {
		return nil
	}
	return FileControllerImpl{
		configuration: configuration,
		fileService:   fileService,
	}
}

func (fileController FileControllerImpl) GetAllFiles(context *gin.Context) {
	files, err := fileController.fileService.GetAllFiles(getUser(context))
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to retrieve files: "+err.Error())
		return
	}
	filesDto := funcs.Maps(files, models.FileToDto).([]models.FileDTO)
	context.JSON(http.StatusOK, filesDto)
}

func (fileController FileControllerImpl) GetAllFileVersions(context *gin.Context) {
	filename := context.Param("file")
	user := getUser(context)
	files, err := fileController.fileService.GetFileVersionsByNameAndOwner(filename, &user)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to retrieve files: "+err.Error())
		return
	}
	filesDto := funcs.Maps(files, models.FileToDto).([]models.FileDTO)
	context.JSON(http.StatusOK, filesDto)
}

func (fileController FileControllerImpl) GetFile(context *gin.Context) {
	filename := context.Param("file")
	user := getUser(context)
	file, err := fileController.fileService.GetLastVersionFileByNameAndOwner(filename, &user)

	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		}
	}
	context.JSON(200, models.FileToDto(file))
}

func (fileController FileControllerImpl) GetFileVersion(context *gin.Context) {
	filename := context.Param("file")
	versionStr := context.Param("version")
	user := getUser(context)
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to parse version number: "+err.Error())
		return
	}
	file, err := fileController.fileService.GetFileByVersion(filename, version, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, fmt.Sprintf(`Unable to find file "%v" version %v: %v`, filename, version, err))
			return
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v" version %v: %v`, filename, version, err))
			return
		}
	}
	context.JSON(200, models.FileToDto(file))
}

/*else {
// ---------- gdrive ----------
gService, err := services.GetGdriveService()
if err != nil {
sendJsonMsg(context, http.StatusInternalServerError, err.Error())
return
}

err = services.UpdateFile(gService, fileModel.GdriveID.String, strconv.FormatInt(fileModel.Id, 10), file)
if err != nil {
sendJsonMsg(context, http.StatusInternalServerError, err.Error())
fmt.Printf("UpdateFile gdrive error: %v", err.Error())
return
}
// ----------------------------

_, err = fileController.fileService.UpdateFile(fileModel.Id, fileModel)
if err != nil {
sendJsonMsg(context, http.StatusInternalServerError, err.Error())
return
}
}
file, err := fileController.fileService.GetFileByVersion(filename, version, &user)
if err != nil {
if err == sql.ErrNoRows {
sendJsonMsg(context, http.StatusNotFound, fmt.Sprintf(`Unable to find file "%v" version %v: %v`, filename, version, err))
return
} else {
sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v" version %v: %v`, filename, version, err))
return
}
}
context.JSON(200, models.FileToDto(file))
}
*/

func (fileController FileControllerImpl) DownloadFile(context *gin.Context) {
	filename := context.Param("file")
	user := getUser(context)

	file, err := fileController.fileService.GetLastVersionFileByNameAndOwner(filename, &user)
	fileController.download(filename, file, err, context)
}

func (fileController FileControllerImpl) DownloadFileVersion(context *gin.Context) {
	filename := context.Param("file")
	versionStr := context.Param("version")
	user := getUser(context)
	version, err := strconv.ParseInt(versionStr, 10, 64)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to parse version number: "+err.Error())
		return
	}

	file, err := fileController.fileService.GetFileByVersion(filename, version, &user)
	fileController.download(filename, file, err, context)
}

func (fileController FileControllerImpl) download(filename string, file models.File, err error, context *gin.Context) {
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		} else {
			sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
			return
		}
	}

	// Local download
	// fileDecrypt, err := fileController.fileService.GetDecryptedLocalFile(file)
	// GDrive download
	fileDecrypt, err := fileController.fileService.GetFileGDrive(file)

	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, fmt.Sprintf(`Unable to find file "%v": %v`, filename, err))
	}

	contentLength, contentType, reader, extraHeaders := fileController.fileService.GetFileStream(fileDecrypt, file)
	context.DataFromReader(http.StatusOK, contentLength, contentType, reader, extraHeaders)
}

func (fileController FileControllerImpl) UploadFile(context *gin.Context) {
	file, header, err := context.Request.FormFile("file")
	permissionsStr, _ := context.GetPostForm("permissions")
	if permissionsStr != "" && len(permissionsStr) != 9 {
		sendJsonMsg(context, http.StatusBadRequest, "Wrong permissions flags (must have 9 characters exactly)")
		return
	}
	if err != nil {
		sendJsonMsg(context, http.StatusBadRequest, err.Error())
		return
	}

	fileModel, err := fileController.fileService.CreateFile(&models.File{
		Name:           header.Filename,
		Owner:          getUser(context),
		PermissionsStr: permissionsStr,
	})
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}

	// Normal upload
	// err = fileController.fileService.SaveFile(file, fileModel)

	// Gdrive upload
	gService, err := services.GetGdriveService()
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}
	fileDrive, err := fileController.fileService.UploadFileGDrive(gService, strconv.FormatInt(fileModel.Id, 10), file)
	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, "Unable to upload file to google drive: "+err.Error())
		return
	}
	err = fileController.fileService.SetGdriveId(fileModel.Id, fileDrive.Id)

	if err != nil {
		sendJsonMsg(context, http.StatusInternalServerError, err.Error())
		return
	}

	context.JSON(http.StatusCreated, models.FileToDto(fileModel))
}

func (fileController FileControllerImpl) DeleteFile(context *gin.Context) {
	filename := context.Param("file")
	user := getUser(context)

	_, err := fileController.fileService.DeleteFile(filename, &user)
	if err != nil {
		if err == sql.ErrNoRows {
			sendJsonMsg(context, http.StatusNotFound, "Unable to find file: "+filename)
		} else {
			sendJsonMsg(context, http.StatusBadRequest, "Unable to delete file: "+err.Error())
		}
		return
	}
	context.Writer.WriteHeader(http.StatusNoContent)
}

func getUser(context *gin.Context) models.User {
	var user models.User
	user.Email = jwt.ExtractClaims(context)["id"].(string)
	return user
}
