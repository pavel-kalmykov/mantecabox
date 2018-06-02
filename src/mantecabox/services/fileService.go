package services

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"

	"mantecabox/dao/factory"
	"mantecabox/dao/interfaces"
	"mantecabox/models"
	"mantecabox/utilities/aes"

	"github.com/go-http-utils/headers"
	"github.com/sirupsen/logrus"
)

var (
	fileDao interfaces.FileDao
	path    = "./files/"
)

type (
	FileService interface {
		GetAllFiles(user models.User) ([]models.File, error)
		CreateFile(file *models.File) (models.File, error)
		DeleteFile(file int64, fileID string) error
		GetFile(filename string, user *models.User) (models.File, error)
		UpdateFile(id int64, file models.File) (models.File, error)
		SaveFile(file multipart.File, uploadedFile models.File) error
		GetDecryptedLocalFile(file models.File) ([]byte, error)
		GetFileStream(fileDecrypt []byte, file models.File) (contentLength int64, contentType string, reader *bytes.Reader, extraHeaders map[string]string)
	}

	FileServiceImpl struct {
		configuration *models.Configuration
	}
)

func init() {
	dao := factory.FileDaoFactory("postgres")
	fileDao = dao
	createDirIfNotExist()
}

func NewFileService(configuration *models.Configuration) FileService {
	if configuration == nil {
		return nil
	}
	return FileServiceImpl{
		configuration: configuration,
	}
}

func (fileService FileServiceImpl) GetAllFiles(user models.User) ([]models.File, error) {
	return fileDao.GetAll(&user)
}

func (fileService FileServiceImpl) CreateFile(file *models.File) (models.File, error) {
	return fileDao.Create(file)
}

func (fileService FileServiceImpl) DeleteFile(file int64, fileID string) error {

	err := fileDao.Delete(file)
	if err != nil {
		return err
	}

	err = os.Remove(path + fileID)
	if err != nil {
		return err
	}

	return err
}

func (fileService FileServiceImpl) GetFile(filename string, user *models.User) (models.File, error) {
	return fileDao.GetByPk(filename, user)
}

func (fileService FileServiceImpl) UpdateFile(id int64, file models.File) (models.File, error) {
	return fileDao.Update(id, &file)
}

func (fileService FileServiceImpl) SaveFile(file multipart.File, uploadedFile models.File) error {
	// Conversión a bytes del fichero
	buf := bytes.NewBuffer(nil)
	if _, err := io.Copy(buf, file); err != nil {
		return err
	}
	encrypted := aes.Encrypt(buf.Bytes())
	// Guardamos el fichero encriptado
	if err := ioutil.WriteFile(path+strconv.FormatInt(uploadedFile.Id, 10), encrypted, 0600); err != nil {
		return err
	}
	return nil
}

func (fileService FileServiceImpl) GetDecryptedLocalFile(file models.File) ([]byte, error) {
	fileEncrypt, err := ioutil.ReadFile(path + strconv.FormatInt(file.Id, 10))
	if err != nil {
		return nil, err
	}

	return aes.Decrypt(fileEncrypt), err
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

func createDirIfNotExist() {
	err := os.MkdirAll(path, 0600)
	if err != nil {
		logrus.Print("Error en la creación del directorio")
		panic(err)
	}
}
