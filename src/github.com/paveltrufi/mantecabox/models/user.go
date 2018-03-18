package models

type User struct {
	TimeStamp
	SoftDelete
	Username string `json:"username"`
	Password string `json:"password"`
}
