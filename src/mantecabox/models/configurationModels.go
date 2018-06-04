package models

type Database struct {
	Engine   string `json:"engine"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type Server struct {
	Host string `json:"host"`
	Port int    `json:"port"`
	Cert string `json:"cert"`
	Key  string `json:"key"`
}

type Mail struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Configuration struct {
	AesKey                    string   `json:"aes_key"`
	TokenTimeout              string   `json:"token_timeout"`
	BlockedLoginTimeLimit     string   `json:"blocked_login_time_limit"`
	VerificationMailTimeLimit string   `json:"verification_mail_time_limit"`
	MaxUnsuccessfulAttempts   int      `json:"max_unsuccessful_attempts"`
	FilesPath                 string   `json:"files_path"`
	UseGDrive                 bool     `json:"use_gdrive"`
	Database                  Database `json:"database"`
	Server                    Server   `json:"server"`
	Mail                      Mail     `json:"mail"`
}
