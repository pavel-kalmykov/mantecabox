package models

import (
	"gopkg.in/guregu/null.v3"
)

type SoftDelete struct {
	DeletedAt null.Time `json:"deleted_at"`
}
