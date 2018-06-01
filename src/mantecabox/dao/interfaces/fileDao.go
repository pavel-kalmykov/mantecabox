package interfaces

import "mantecabox/models"

type FileDao interface {
	GetAll() ([]models.File, error)
	GetByPk(filename string, user *models.User) (models.File, error)
	Create(f *models.File) (models.File, error)
	Update(id int64, f *models.File) (models.File, error)
	Delete(id int64) error
}
