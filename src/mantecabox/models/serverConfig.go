package models

type ServerConfig struct {
	Port         string `json:"port"`
	Certificates `json:"certificates"`
}
