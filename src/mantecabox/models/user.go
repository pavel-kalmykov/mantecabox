package models

import (
	"gopkg.in/guregu/null.v3"
)

type Credentials struct {
	Email         string      `json:"email"`
	Password      string      `json:"password"`
	TwoFactorAuth null.String `json:"two_factor_auth"`
	TwoFactorTime null.Time   `json:"two_factor_time"`
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
