package models

import "github.com/lib/pq"

type TimeStamp struct {
	CreatedAt pq.NullTime `json:"created_at"`
	UpdatedAt pq.NullTime `json:"updated_at"`
}
