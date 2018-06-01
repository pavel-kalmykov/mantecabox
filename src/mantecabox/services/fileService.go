package services

import (
	"mantecabox/config"
	"mantecabox/dao/factory"
	"mantecabox/dao/interfaces"
	"mantecabox/models"
)

var (
	fileDao interfaces.FileDao
)

func init() {
	dao := factory.FileDaoFactory(config.GetServerConf().Engine)
	fileDao = dao
}

func CreateFile (file *models.File) (models.File, error) {
	return fileDao.Create(file)
}

func DeleteFile (file int64) error {
	return fileDao.Delete(file)
}

func DownloadFile (filename string, user *models.User) (models.File, error) {
	return fileDao.GetByPk(filename, user)
}
