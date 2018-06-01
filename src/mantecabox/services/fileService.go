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

	"github.com/sirupsen/logrus"
)

var (
	fileDao interfaces.FileDao
	Path    = "./files/"
)

func init() {
	dao := factory.FileDaoFactory(config.GetServerConf().Engine)
	fileDao = dao
	CreateDirIfNotExist()
}

func CreateDirIfNotExist() {
	err := os.MkdirAll(Path, 0600)
	if err != nil {
		logrus.Print("Error en la creación del directorio")
		panic(err)
	}
}

func CreateFile (file *models.File) (models.File, error) {
	return fileDao.Create(file)
}

func DeleteFile (file int64) error {
	return fileDao.Delete(file)
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
	if err := ioutil.WriteFile(Path+strconv.FormatInt(uploadedFile.Id, 10), encrypted, 0600); err != nil {
		return err
	}
	return nil
}
