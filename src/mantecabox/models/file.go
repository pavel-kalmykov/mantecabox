package models

type File struct {
	Id    int64  `json:"id"`
	TimeStamp
	SoftDelete
	Name  string `json:"name"`
	Owner User   `json:"owner"`
}

type FileDTO struct {
	TimeStamp
	Name string `json:"name"`
}

func FileToDto(file File) FileDTO {
	return FileDTO{
		TimeStamp: file.TimeStamp,
		Name:      file.Name,
	}
}
