package services

import (
	"errors"

	"mantecabox/dao/factory"
	"mantecabox/models"

	"github.com/benashford/go-func"
)

const (
	MaxUnsuccessfulAttempts = 3
)

var (
	loginAttemptDao    = factory.LoginAttemptFactory(configuration.Engine)
	TooManyAttemptsErr = errors.New("too many login attemtps")
)

func ProcessLoginAttempt(attempt *models.LoginAttempt) error {
	_, err := loginAttemptDao.Create(attempt)
	if err != nil {
		return err
	}
	attempts, err := loginAttemptDao.GetLastNByUser(attempt.User.Email, MaxUnsuccessfulAttempts)
	if err != nil {
		return err
	}
	// First, we look if the last N attempts are all unsuccessful
	unsuccessfulAttempts := funcs.Filters(attempts, func(a models.LoginAttempt) bool {
		return !a.Successful
	}).([]models.LoginAttempt)
	if len(unsuccessfulAttempts) == MaxUnsuccessfulAttempts {
		sendSuspiciousActivityReport(unsuccessfulAttempts)
		return TooManyAttemptsErr
	}
	// Then, we look if similar attempt data were added before or if this login occurred in a new device or place
	similarAttempts, err := loginAttemptDao.GetSimilarAttempts(attempt)
	if err != nil {
		return err
	}
	if len(similarAttempts) == 0 {
		return sendNewRegisteredDeviceActivity(attempt)
	}
	return nil
}

func sendNewRegisteredDeviceActivity(attempt *models.LoginAttempt) error {

	return nil
}

func sendSuspiciousActivityReport(unsuccessfulAttempts []models.LoginAttempt) error {
	return nil
}
