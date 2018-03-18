package models

import "github.com/lib/pq"

type SoftDelete struct {
	DeletedAt pq.NullTime `json:"deleted_at"`
}
