package dao

import (
	"database/sql"
	"errors"

	"mantecabox/models"

	"github.com/sirupsen/logrus"
)

const (
	getAllUsersQuery = "SELECT * FROM users WHERE deleted_at IS NULL"
	getUserByPkQuery = "SELECT * FROM users WHERE deleted_at IS NULL AND email = $1"
	insertUserQuery  = "INSERT INTO users(email,password) VALUES($1,$2) RETURNING *;"
	updateUserQuery  = "UPDATE users SET email=$1, password=$2, two_factor_auth=$3 WHERE email=$4 RETURNING *"
	deleteUserQuery  = "UPDATE users SET deleted_at = NOW() WHERE email = $1"
)

var daoLog = logrus.WithFields(logrus.Fields{"package": "postgres"})

type UserPgDao struct {
}

func (dao UserPgDao) GetAll() ([]models.User, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		users := make([]models.User, 0)
		rows, err := db.Query(getAllUsersQuery)
		if err != nil {
			daoLog.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
			return nil, err
		}

		for rows.Next() {
			var user models.User
			err := scanUserRow(rows, &user)
			if err != nil {
				daoLog.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
				return nil, err
			}
			users = append(users, user)
		}

		daoLog.Debug("Queried", len(users), "users")
		return users, err
	})
	return res.([]models.User), err
}

func (dao UserPgDao) GetByPk(email string) (models.User, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		user := models.User{}
		row := db.QueryRow(getUserByPkQuery, email)
		err := scanUserRow(row, &user)
		if err != nil {
			daoLog.Debug("Unable to execute UserPgDao.GetByPk(email string) query. Reason:", err)
		} else {
			daoLog.Debug("Retrieved user", user)
		}
		return user, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Create(user *models.User) (models.User, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var createdUser models.User
		row := db.QueryRow(insertUserQuery,
			user.Email, user.Password)
		err := scanUserRow(row, &createdUser)
		if err != nil {
			daoLog.Info("Unable to execute UserPgDao.Create(user models.User) query. Reason:", err)
		} else {
			daoLog.Debug("Created user", createdUser)
		}
		return createdUser, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Update(email string, user *models.User) (models.User, error) {
	res, err := withDb(func(db *sql.DB) (interface{}, error) {
		var updatedUser models.User
		row := db.QueryRow(updateUserQuery,
			user.Email, user.Password, user.TwoFactorAuth, email)
		err := scanUserRow(row, &updatedUser)
		if err != nil {
			daoLog.Info("Unable to execute UserPgDao.Update(email string, user models.User) query. Reason:", err)
		} else {
			daoLog.Debug("Updated user", updatedUser)
		}
		return updatedUser, err
	})
	return res.(models.User), err
}

func (dao UserPgDao) Delete(email string) error {
	_, err := withDb(func(db *sql.DB) (interface{}, error) {
		result, err := db.Exec(deleteUserQuery, email)
		if err != nil {
			daoLog.Info("Unable to execute UserPgDao.Delete(email string) query. Reason:", err)
		} else {
			var rowsAffected int64
			rowsAffected, err = result.RowsAffected()
			if err != nil {
				daoLog.Info("Some error occured during deleting:", err)
			} else {
				switch {
				case rowsAffected == 0:
					err = sql.ErrNoRows
				case rowsAffected > 1:
					err = errors.New("more than one deleted")
				}
				if err != nil {
					daoLog.Debug("Unable to delete user with email \""+email+"\" correctly. Reason:", err)
				} else {
					daoLog.Debug("User with email \"" + email + "\" successfully deleted")
				}
			}
		}
		return nil, err
	})
	return err
}

type polimorphicScanner interface {
	Scan(dest ...interface{}) error
}

func scanUserRow(scanner polimorphicScanner, user *models.User) error {
	err := scanner.Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Email,
		&user.Password,
		&user.TwoFactorAuth,
		&user.TwoFactorTime)
	return err
}
