package models

type Credentials struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type User struct {
	TimeStamp
	SoftDelete
	Credentials
}

type UserWithFiles struct {
	User
	Files []File `json:"files"`
}
