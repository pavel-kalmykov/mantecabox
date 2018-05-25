package models

import (
	"database/sql"

	"github.com/lib/pq"
)

type Credentials struct {
	Email         string         `json:"email"`
	Password      string         `json:"password"`
	TwoFactorAuth sql.NullString `json:"two_factor_auth"`
	TwoFactorTime pq.NullTime    `json:"two_factor_time"`
}

type User struct {
	TimeStamp
	SoftDelete
	Credentials
}

type UserDto struct {
	Email string `json:"email"`
}

type UserWithFiles struct {
	User
	Files []File `json:"files"`
}
