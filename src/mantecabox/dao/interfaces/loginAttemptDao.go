package interfaces

import (
	"mantecabox/models"
)

type LoginAttempDao interface {
	GetByUser(email string) ([]models.LoginAttempt, error)
	GetLastNByUser(email string, n int) ([]models.LoginAttempt, error)
	GetSimilarAttempts(attempt *models.LoginAttempt) ([]models.LoginAttempt, error)
	Create(attempt *models.LoginAttempt) (models.LoginAttempt, error)
}
