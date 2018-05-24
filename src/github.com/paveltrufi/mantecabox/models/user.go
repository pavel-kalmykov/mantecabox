package models

type Credentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
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
