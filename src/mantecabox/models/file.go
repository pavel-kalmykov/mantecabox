package models

type File struct {
	Id    int64  `json:"id"`
	TimeStamp
	SoftDelete
	Name  string `json:"name"`
	Owner User   `json:"owner"`
}

type FileDTO struct {
	Name  string `json:"name"`
	Owner string `json:"owner"`
}
