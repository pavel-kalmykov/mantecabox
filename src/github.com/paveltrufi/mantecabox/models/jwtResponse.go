package models

import (
	"time"
)

type JwtResponse struct {
	Code   int       `json:"code"`
	Token  string    `json:"token"`
	Expire time.Time `json:"expire"`
}
