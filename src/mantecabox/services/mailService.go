package services

import (
	"fmt"

	"mantecabox/logs"
	"mantecabox/models"

	"github.com/badoux/checkmail"
	"github.com/hako/durafmt"
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
	durationStr, err := durafmt.ParseString(mailService.configuration.VerificationMailTimeLimit)
	if err != nil {
		logs.ServicesLog.Fatalf("unable to parse mail's verification limit configuration value: %v" + err.Error())
	}
	return mailService.SendMail(toEmail, fmt.Sprintf("Hello. Your security code is M-<strong>%v</strong>. It will expire in %v", code, durationStr))
}

func (mailService MailServiceImpl) SendMail(toEmail, bodyMessage string) error {
	mailConf := mailService.configuration.Mail
	m := gomail.NewMessage()
	m.SetHeader("From", mailConf.Username)
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "Mantecabox Backup")
	m.SetBody("text/html", bodyMessage)
	return gomail.
		NewDialer(mailConf.Host, mailConf.Port, mailConf.Username, mailConf.Password).
		DialAndSend(m)
}
