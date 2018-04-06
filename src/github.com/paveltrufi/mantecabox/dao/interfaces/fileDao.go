package interfaces

import "github.com/paveltrufi/mantecabox/models"

type FileDao interface {
	GetAll() ([]models.File, error)
	GetByPk(id int64) (models.File, error)
	Create(f *models.File) (models.File, error)
	Update(id int64, f *models.File) (models.File, error)
	Delete(id int64) error
}
