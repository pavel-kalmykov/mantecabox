package services

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"mantecabox/dao"
	"mantecabox/models"
	"mantecabox/utilities"

	"github.com/go-http-utils/headers"
	"github.com/sirupsen/logrus"
)

type (
	FileService interface {
		GetAllFiles(user models.User) ([]models.File, error)
		GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error)
		GetFileStream(fileDecrypt []byte, file models.File) (contentLength int64, contentType string, reader *bytes.Reader, extraHeaders map[string]string)
		GetDecryptedLocalFile(file models.File) ([]byte, error)
		CreateFile(file *models.File) (models.File, error)
		SaveFile(file multipart.File, uploadedFile models.File) error
		DeleteFile(filename string, user *models.User) error
		createDirIfNotExists()
	}

	FileServiceImpl struct {
		configuration *models.Configuration
		fileDao       dao.FileDao
		aesCipher     utilities.AesCTRCipher
	}
)

func NewFileService(configuration *models.Configuration) FileService {
	if configuration == nil {
		return nil
	}
	if configuration.FilesPath == "" {
		configuration.FilesPath = "files"
	}
	// Maybe the config path didn't ended with folder's slash, so we add it
	if configuration.FilesPath[len(configuration.FilesPath)-1] != '/' {
		configuration.FilesPath += "/"
	}
	fileServiceImpl := FileServiceImpl{
		configuration: configuration,
		fileDao:       dao.FileDaoFactory(configuration.Database.Engine),
		aesCipher:     utilities.NewAesCTRCipher(configuration.AesKey),
	}
	fileServiceImpl.createDirIfNotExists()
	return fileServiceImpl
}

func (fileService FileServiceImpl) GetAllFiles(user models.User) ([]models.File, error) {
	return fileService.fileDao.GetAllByOwner(&user)
}

func (fileService FileServiceImpl) GetLastVersionFileByNameAndOwner(filename string, user *models.User) (models.File, error) {
	return fileService.fileDao.GetLastVersionFileByNameAndOwner(filename, user)
}

func (fileService FileServiceImpl) GetDecryptedLocalFile(file models.File) ([]byte, error) {
	fileEncrypt, err := ioutil.ReadFile(fileService.configuration.FilesPath + strconv.FormatInt(file.Id, 10))
	if err != nil {
		return nil, err
	}

	return fileService.aesCipher.Decrypt(fileEncrypt), err
}

func (fileService FileServiceImpl) GetFileStream(fileDecrypt []byte, file models.File) (contentLength int64, contentType string, reader *bytes.Reader, extraHeaders map[string]string) {
	reader = bytes.NewReader(fileDecrypt)
	contentLength = reader.Size()
	contentType = http.DetectContentType(fileDecrypt)

	extraHeaders = map[string]string{
		headers.ContentDisposition: `attachment; filename="` + file.Name + `"`,
	}

	return
}

func (fileService FileServiceImpl) CreateFile(file *models.File) (models.File, error) {
	return fileService.fileDao.Create(file)
}

func (fileService FileServiceImpl) SaveFile(file multipart.File, uploadedFile models.File) error {
	// Conversión a bytes del fichero
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return err
	}
	encrypted := fileService.aesCipher.Encrypt(buf.Bytes())
	// Guardamos el fichero encriptado
	if err := ioutil.WriteFile(fileService.configuration.FilesPath+strconv.FormatInt(uploadedFile.Id, 10), encrypted, 0600); err != nil {
		return err
	}
	return nil
}

func (fileService FileServiceImpl) DeleteFile(filename string, user *models.User) error {
	file, err := fileService.fileDao.GetLastVersionFileByNameAndOwner(filename, user)
	if err != nil {
		return err
	}
	err = fileService.fileDao.Delete(filename, user)
	if err != nil {
		return err
	}

	err = os.Remove(fileService.configuration.FilesPath + strconv.FormatInt(file.Id, 10))
	if err != nil {
		return err
	}

	return err
}

func (fileService FileServiceImpl) createDirIfNotExists() {
	err := os.MkdirAll(fileService.configuration.FilesPath, 0700)
	if err != nil {
		logrus.Print("Error creating file's directory: " + err.Error())
		panic(err)
	}
}
