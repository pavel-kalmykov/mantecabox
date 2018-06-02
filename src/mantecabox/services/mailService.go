package services

import (
	"fmt"

	"mantecabox/models"

	"github.com/badoux/checkmail"
	"gopkg.in/gomail.v2"
)

type (
	MailService interface {
		Send2FAEmail(toEmail, code string) error
		SendMail(toEmail, bodyMessage string) error
	}

	MailServiceImpl struct {
		configuration *models.Configuration
	}
)

func NewMailService(configuration *models.Configuration) MailService {
	if configuration == nil {
		return nil
	}
	return MailServiceImpl{
		configuration: configuration,
	}
}

func (mailService MailServiceImpl) Send2FAEmail(toEmail, code string) error {
	if code == "" {
		return Empty2FACodeError
	}
	if err := checkmail.ValidateHost(toEmail); err != nil {
		return err
	}
	return mailService.SendMail(toEmail, fmt.Sprintf("Hello. Your security code is M-<b>%v</b>. It will expire in 5 minutes", code))
}

func (mailService MailServiceImpl) SendMail(toEmail, bodyMessage string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", "mantecabox@gmail.com")
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Mantecabox Backup")
	m.SetBody("text/html", bodyMessage)
	return gomail.
		NewDialer("smtp.gmail.com", 587, "mantecabox@gmail.com", "ElPutoPavel").
		DialAndSend(m)
}
