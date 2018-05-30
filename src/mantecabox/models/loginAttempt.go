package models

import (
	"time"

	"gopkg.in/guregu/null.v3"
)

type LoginAttempt struct {
	Id         int64       `json:"id"`
	CreatedAt  time.Time   `json:"created_at"`
	User       User        `json:"user"`
	UserAgent  null.String `json:"user_agent"`
	IP         null.String `json:"ip_v4"`
	Successful bool        `json:"successful"`
}
