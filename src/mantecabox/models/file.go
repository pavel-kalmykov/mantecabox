package models

import (
	"gopkg.in/guregu/null.v3"
)

type File struct {
	Id int64 `json:"id"`
	TimeStamp
	SoftDelete
	Name             string      `json:"name"`
	Owner            User        `json:"owner"`
	Group            null.String `json:"group"`
	UserReadable     bool        `json:"user_readable"`
	UserWritable     bool        `json:"user_writable"`
	UserExecutable   bool        `json:"user_executable"`
	GroupReadable    bool        `json:"group_readable"`
	GroupWritable    bool        `json:"group_writable"`
	GroupExecutable  bool        `json:"group_executable"`
	OtherReadable    bool        `json:"other_readable"`
	OtherWritable    bool        `json:"other_writable"`
	OtherExecutable  bool        `json:"other_executable"`
	PlatformCreation null.String `json:"platform_creation"`
}
