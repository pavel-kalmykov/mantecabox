package services

import (
	"bytes"
	"io"
	"io/ioutil"
	"mime/multipart"
	"os"
	"strconv"

	"mantecabox/config"
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

func init() {
	dao := factory.FileDaoFactory(config.GetServerConf().Engine)
	fileDao = dao
	CreateDirIfNotExist()
}

func CreateDirIfNotExist() {
	err := os.MkdirAll(path, 0600)
	if err != nil {
		logrus.Print("Error en la creación del directorio")
		panic(err)
	}
}

func GetAllFiles(user models.User) ([]models.File, error) {
	return fileDao.GetAll(&user)
}

func CreateFile(file *models.File) (models.File, error) {
	return fileDao.Create(file)
}

func DeleteFile(file int64, fileID string) error {

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

func GetFile(filename string, user *models.User) (models.File, error) {
	return fileDao.GetByPk(filename, user)
}

func UpdateFile(id int64, file models.File) (models.File, error) {
	return fileDao.Update(id, &file)
}

func SaveFile(file multipart.File, uploadedFile models.File) error {
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

func GetDecryptedLocalFile(file models.File) ([]byte, error) {
	fileEncrypt, err := ioutil.ReadFile(path + strconv.FormatInt(file.Id, 10))
	if err != nil {
		return nil, err
	}

	return aes.Decrypt(fileEncrypt), err
}

func GetFileStream(fileDecrypt []byte, file models.File) (contentLength int64, contentType string, reader *bytes.Reader, extraHeaders map[string]string) {
	reader = bytes.NewReader(fileDecrypt)
	contentLength = reader.Size()
	contentType = "application/octet-stream"

	extraHeaders = map[string]string{
		headers.ContentDisposition: `attachment; filename="` + file.Name + `"`,
	}

	return
}
