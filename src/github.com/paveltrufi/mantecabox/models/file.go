package models

import "database/sql"

type File struct {
	Id int64 `json:"id"`
	TimeStamp
	SoftDelete
	Name             string         `json:"name"`
	Owner            User           `json:"owner"`
	Group            sql.NullString `json:"group"`
	UserReadable     bool           `json:"user_readable"`
	UserWritable     bool           `json:"user_writable"`
	UserExecutable   bool           `json:"user_executable"`
	GroupReadable    bool           `json:"group_readable"`
	GroupWritable    bool           `json:"group_writable"`
	GroupExecutable  bool           `json:"group_executable"`
	OtherReadable    bool           `json:"other_readable"`
	OtherWritable    bool           `json:"other_writable"`
	OtherExecutable  bool           `json:"other_executable"`
	PlatformCreation sql.NullString `json:"platform_creation"`
}
