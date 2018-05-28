package postgres

import (
	"database/sql"
	"errors"

	"mantecabox/database"
	"mantecabox/models"

	log "github.com/alexrudd/go-logger"
)

type UserPgDao struct {
}

func (dao UserPgDao) GetAll() ([]models.User, error) {
	users := make([]models.User, 0)
	db, err := database.GetPgDb()
	if err != nil {
		log.Fatal("Unable to connnect with database")
	}
	defer db.Close()

	rows, err := db.Query("SELECT * FROM users WHERE deleted_at IS NULL")
	if err != nil {
		log.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
		return nil, err
	}

	for rows.Next() {
		var user models.User
		err := rows.Scan(
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
			&user.Email,
			&user.Password,
			&user.TwoFactorAuth,
			&user.TwoFactorTime)
		if err != nil {
			log.Info("Unable to execute UserPgDao.GetAll() query. Reason:", err)
			return nil, err
		}
		users = append(users, user)
	}

	log.Debug("Queried", len(users), "users")
	return users, err
}

func (dao UserPgDao) GetByPk(email string) (models.User, error) {
	user := models.User{}
	db, err := database.GetPgDb()
	if err != nil {
		log.Fatal("Unable to connnect with database")
	}
	defer db.Close()

	err = db.QueryRow("SELECT * FROM users WHERE email = $1 AND deleted_at IS NULL", email).Scan(
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
		&user.Email,
		&user.Password,
		&user.TwoFactorAuth,
		&user.TwoFactorTime)

	if err != nil {
		log.Debug("Unable to execute UserPgDao.GetByPk(email string) query. Reason:", err)
	} else {
		log.Debug("Retrieved user", user)
	}
	return user, err
}

func (dao UserPgDao) Create(user *models.User) (models.User, error) {
	db, err := database.GetPgDb()
	if err != nil {
		log.Fatal("Unable to connnect with database")
	}
	defer db.Close()

	var createdUser models.User
	err = db.QueryRow("INSERT INTO users(email,password) VALUES($1,$2) RETURNING *;",
		user.Email, user.Password).Scan(
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
		&createdUser.DeletedAt,
		&createdUser.Email,
		&createdUser.Password,
		&createdUser.TwoFactorAuth,
		&createdUser.TwoFactorTime)

	if err != nil {
		log.Info("Unable to execute UserPgDao.Create(user models.User) query. Reason:", err)
	} else {
		log.Debug("Created user", createdUser)
	}
	return createdUser, err
}

func (dao UserPgDao) Update(email string, user *models.User) (models.User, error) {
	db, err := database.GetPgDb()
	if err != nil {
		log.Fatal("Unable to connnect with database")
	}
	defer db.Close()

	var updatedUser models.User
	err = db.QueryRow("UPDATE users SET email=$1, password=$2, two_factor_auth=$3 WHERE email=$4 RETURNING *",
		user.Email, user.Password, user.TwoFactorAuth, email).Scan(
		&updatedUser.CreatedAt,
		&updatedUser.UpdatedAt,
		&updatedUser.DeletedAt,
		&updatedUser.Email,
		&updatedUser.Password,
		&updatedUser.TwoFactorAuth,
		&updatedUser.TwoFactorTime)

	if err != nil {
		log.Info("Unable to execute UserPgDao.Update(email string, user models.User) query. Reason:", err)
	} else {
		log.Debug("Updated user", updatedUser)
	}
	return updatedUser, err
}

func (dao UserPgDao) Delete(email string) error {
	db, err := database.GetPgDb()
	if err != nil {
		log.Fatal("Unable to connnect with database")
	}
	defer db.Close()

	result, err := db.Exec("DELETE FROM users WHERE email = $1", email)
	if err != nil {
		log.Info("Unable to execute UserPgDao.Delete(email string) query. Reason:", err)
	} else {
		var rowsAffected int64
		rowsAffected, err = result.RowsAffected()
		if err != nil {
			log.Info("Some error occured during deleting:", err)
		} else {
			switch {
			case rowsAffected == 0:
				err = sql.ErrNoRows
			case rowsAffected > 1:
				err = errors.New("more than one deleted")
			}
			if err != nil {
				log.Debug("Unable to delete user with email \""+email+"\" correctly. Reason:", err)
			} else {
				log.Debug("User with email \"" + email + "\" successfully deleted")
			}
		}
	}
	return err
}
