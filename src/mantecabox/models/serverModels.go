package models

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type JwtResponse struct {
	Code   int       `json:"code"`
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}

type LoginAttempt struct {
	Id         int64       `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	User       User        `json:"user"`
	UserAgent  null.String `json:"user_agent"`
	IP         null.String `json:"ip"`
	Successful bool        `json:"successful"`
}

type ServerError struct {
	Message string `json:"message"`
}
