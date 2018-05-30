package postgres

import (
	"database/sql"

	"mantecabox/database"
	"mantecabox/models"

	"github.com/sirupsen/logrus"
)

type LoginAttemptPgDao struct {
}

func withDb(f func(db *sql.DB) (models.LoginAttempt, error)) (models.LoginAttempt, error) {
	db, err := database.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
		return models.LoginAttempt{}, err
	}
	defer db.Close()
	return f(db)
}

func withDbArray(f func(db *sql.DB) ([]models.LoginAttempt, error)) ([]models.LoginAttempt, error) {
	db, err := database.GetPgDb()
	if err != nil {
		logrus.Fatal("Unable to connnect with database: " + err.Error())
		return nil, err
	}
	defer db.Close()
	return f(db)
}

func scanLoginAttemptWithNestedUser(rows *sql.Rows) ([]models.LoginAttempt, error) {
	attempts := make([]models.LoginAttempt, 0)
	for rows.Next() {
		var attempt models.LoginAttempt
		var user string
		err := rows.Scan(&attempt.Id,
			&attempt.CreatedAt,
			&user,
			&attempt.UserAgent,
			&attempt.IPv4,
			&attempt.IPv6,
			&attempt.Successful,
			&attempt.User.CreatedAt,
			&attempt.User.UpdatedAt,
			&attempt.User.DeletedAt,
			&attempt.User.Email,
			&attempt.User.Password,
			&attempt.User.TwoFactorAuth,
			&attempt.User.TwoFactorTime,
		)
		if err != nil {
			logrus.Info("Unable to execute LoginAttemptPgDao.scanLoginAttemptWithNestedUser() scan. Reason:", err)
			return nil, err
		}
		attempts = append(attempts, attempt)
	}

	logrus.Debug("Queried", len(attempts), "login attempts")
	return attempts, nil
}

func (dao LoginAttemptPgDao) GetByUser(email string) ([]models.LoginAttempt, error) {
	return withDbArray(func(db *sql.DB) ([]models.LoginAttempt, error) {
		rows, err := db.Query(`SELECT
  la.*,
  u.*
FROM login_attempts la
  JOIN users u ON la."user" = u.email
WHERE u.deleted_at IS NULL AND la."user" = $1`, email)
		if err != nil {
			logrus.Info("Unable to execute LoginAttemptPgDao.GetByUser() query. Reason:", err)
			return nil, err
		}
		return scanLoginAttemptWithNestedUser(rows)
	})
}

func (dao LoginAttemptPgDao) GetLastNByUser(email string, n int) ([]models.LoginAttempt, error) {
	if n < 0 {
		return dao.GetByUser(email)
	}
	return withDbArray(func(db *sql.DB) ([]models.LoginAttempt, error) {
		rows, err := db.Query(`SELECT *
FROM (SELECT
        la.*,
        u.*
      FROM login_attempts la
        JOIN users u ON la."user" = u.email
      WHERE u.deleted_at IS NULL AND la."user" = $1
      ORDER BY la.id DESC
      LIMIT $2) as reversed
ORDER BY reversed.id`, email, n)
		if err != nil {
			logrus.Info("Unable to execute LoginAttemptPgDao.GetLastNByUser() query. Reason:", err)
			return nil, err
		}
		return scanLoginAttemptWithNestedUser(rows)
	})
}

func (dao LoginAttemptPgDao) Create(attempt *models.LoginAttempt) (models.LoginAttempt, error) {
	return withDb(func(db *sql.DB) (models.LoginAttempt, error) {
		var createdAttempt models.LoginAttempt
		err := db.QueryRow(`INSERT INTO login_attempts ("user", user_agent, ipv4, ipv6, successful) VALUES ($1, $2, $3, $4, $5)
RETURNING *;`, attempt.User.Email, attempt.UserAgent, attempt.IPv4, attempt.IPv6, attempt.Successful).
			Scan(&createdAttempt.Id, &createdAttempt.CreatedAt, &createdAttempt.User.Email,
				&createdAttempt.UserAgent, &createdAttempt.IPv4, &createdAttempt.IPv6, &createdAttempt.Successful)
		if err != nil {
			logrus.Info("Unable to execute FilePgDao.Create(file models.File) query. Reason:", err)
			return createdAttempt, err
		} else {
			logrus.Debug("Created file: ", createdAttempt)
		}
		owner, err := UserPgDao{}.GetByPk(createdAttempt.User.Email)
		createdAttempt.User = owner
		return createdAttempt, err
	})
}

func (dao LoginAttemptPgDao) GetSimilarAttempts(attempt *models.LoginAttempt) ([]models.LoginAttempt, error) {
	return withDbArray(func(db *sql.DB) ([]models.LoginAttempt, error) {
		rows, err := db.Query(`SELECT
		la.*,
		u.*
		FROM login_attempts la
		JOIN users u ON la."user" = u.email
		WHERE u.deleted_at IS NULL
		AND la."user" = $1
		AND la.user_agent = $2
		AND la.ipv4 = $3
		AND la.ipv6 = $4;`, attempt.User.Email, attempt.UserAgent, attempt.IPv4, attempt.IPv6)
		if err != nil {
			logrus.Info("Unable to execute LoginAttemptPgDao.GetSimilarAttempts() query. Reason:", err)
			return nil, err
		}
		return scanLoginAttemptWithNestedUser(rows)
	})
}
