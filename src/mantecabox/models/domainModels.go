package models

import (
	"gopkg.in/guregu/null.v3"
)

type TimeStamp struct {
	CreatedAt null.Time `json:"created_at"`
	UpdatedAt null.Time `json:"updated_at"`
}

type SoftDelete struct {
	DeletedAt null.Time `json:"deleted_at"`
}

type Credentials struct {
	Email         string      `json:"email"`
	Password      string      `json:"password"`
	TwoFactorAuth null.String `json:"two_factor_auth"`
	TwoFactorTime null.Time   `json:"two_factor_time"`
}

type User struct {
	TimeStamp
	SoftDelete
	Credentials
}

type UserDto struct {
	Email string `json:"email"`
}

type File struct {
	Id int64 `json:"id"`
	TimeStamp
	SoftDelete
	Name  string `json:"name"`
	Owner User   `json:"owner"`
	GdriveID null.String `json:"gdrive_id"`
	PermissionsStr string `json:"permissions"`
}

type FileDTO struct {
	Id int64 `json:"id"`
	TimeStamp
	Name           string `json:"name"`
	PermissionsStr string `json:"permissions"`
}

func FileToDto(file File) FileDTO {
	return FileDTO{
		Id:             file.Id,
		TimeStamp:      file.TimeStamp,
		Name:           file.Name,
		PermissionsStr: file.PermissionsStr,
	}
}
