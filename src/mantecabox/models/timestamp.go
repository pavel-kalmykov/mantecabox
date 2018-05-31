package models

import (
	"gopkg.in/guregu/null.v3"
)

type TimeStamp struct {
	CreatedAt null.Time `json:"created_at"`
	UpdatedAt null.Time `json:"updated_at"`
}
