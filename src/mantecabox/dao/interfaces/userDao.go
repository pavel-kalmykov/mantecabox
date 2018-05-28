package interfaces

import "mantecabox/models"

type UserDao interface {
	GetAll() ([]models.User, error)
	GetByPk(username string) (models.User, error)
	Create(u *models.User) (models.User, error)
	Update(username string, u *models.User) (models.User, error)
	Delete(username string) error
}
