package services

import (
	"errors"
	"fmt"
	"net"
	"time"

	"mantecabox/dao/factory"
	"mantecabox/models"

	"github.com/benashford/go-func"
	"github.com/oschwald/geoip2-golang"
	"github.com/tobie/ua-parser/go/uaparser"
)

const (
	MaxUnsuccessfulAttempts = 3
)

var (
	loginAttemptDao    = factory.LoginAttemptFactory(configuration.Engine)
	TooManyAttemptsErr = errors.New("too many unsuccessful login attemtps")
	timeLimit          = time.Minute * 5
)

func ProcessLoginAttempt(attempt *models.LoginAttempt) error {
	createdAttempt, err := loginAttemptDao.Create(attempt)
	if err != nil {
		return err
	}
	attempts, err := loginAttemptDao.GetLastNByUser(createdAttempt.User.Email, MaxUnsuccessfulAttempts+1)
	if err != nil {
		return err
	}
	// First, we look if the last N attempts are all unsuccessful
	unsuccessfulAttempts := funcs.Filters(attempts, func(a models.LoginAttempt) bool {
		return !a.Successful
	}).([]models.LoginAttempt)
	if len(unsuccessfulAttempts) >= MaxUnsuccessfulAttempts && len(attempts)-len(unsuccessfulAttempts) == 0 {
		go func() {
			sendSuspiciousActivityReport(&createdAttempt) // Very slow, run in background
		}()
		return TooManyAttemptsErr
	}
	if timeDiff := attempts[len(attempts)-1].CreatedAt.Sub(attempts[0].CreatedAt); len(unsuccessfulAttempts) == MaxUnsuccessfulAttempts && timeDiff < timeLimit {
		return errors.New(fmt.Sprintf("Login for user %v blocked for the next %.2f minutes", createdAttempt.User.Email, (timeLimit - timeDiff).Minutes()))
	}
	// Then, we look if similar attempt data were added before or if this login occurred in a new device or place
	similarAttempts, err := loginAttemptDao.GetSimilarAttempts(&createdAttempt)
	if err != nil {
		return err
	}
	if len(similarAttempts) == 1 {
		go func() {
			sendNewRegisteredDeviceActivity(&createdAttempt) // Very slow, run in background
		}()
	}
	return nil
}

func sendNewRegisteredDeviceActivity(attempt *models.LoginAttempt) error {
	messageBody := fmt.Sprintf("<strong>A new device has been registered in your account (%v)</strong><br>", attempt.User.Email)
	msg, err := formatAttempt(attempt)
	if err != nil {
		return err
	}
	messageBody += msg
	return SendMail(attempt.User.Email, messageBody)
}

func sendSuspiciousActivityReport(unsuccessfulAttempt *models.LoginAttempt) error {
	email := unsuccessfulAttempt.User.Email
	messageBody := fmt.Sprintf("<strong>We have detected a suspiciuous activity in your account (%v)</strong><br>", email)
	msg, err := formatAttempt(unsuccessfulAttempt)
	if err != nil {
		return err
	}
	messageBody += msg
	return SendMail(email, messageBody)
}

func formatAttempt(attempt *models.LoginAttempt) (string, error) {
	messageBody := fmt.Sprintf("On %v<br>", attempt.CreatedAt.Format(time.RFC1123))

	if !attempt.IP.IsZero() {
		messageBody += fmt.Sprintf("IP: %v<br>", attempt.IP.ValueOrZero())
	}

	parser, err := uaparser.New("regexes.yaml")
	if err != nil {
		return "", err
	}
	if !attempt.UserAgent.IsZero() {
		client := parser.Parse(attempt.UserAgent.ValueOrZero())
		messageBody += fmt.Sprintf("Browser: %v<br>", client.UserAgent.ToString())
		messageBody += fmt.Sprintf("OS: %v<br>", client.Os.ToString())
		messageBody += fmt.Sprintf("Device: %v<br>", client.Device.ToString())
	}

	if !attempt.IP.IsZero() {
		db, err := geoip2.Open("GeoLite2-City.mmdb")
		if err != nil {
			return "", err
		}
		defer db.Close()
		ip := net.ParseIP(attempt.IP.ValueOrZero())
		record, err := db.City(ip)
		if err != nil {
			return "", err
		}
		if record.Location.Latitude != 0 && record.Location.Longitude != 0 {
			region := ""
			if len(record.Subdivisions) > 0 {
				region = ", " + record.Subdivisions[0].Names["en"] + " "
			}
			messageBody += fmt.Sprintf(`Location: %v %v(%v). <a href="https://www.google.es/maps?q=%v,%v">See in maps</a><br>`,
				record.City.Names["en"],
				region,
				record.Country.Names["en"],
				record.Location.Latitude,
				record.Location.Longitude)
		}
	}
	return messageBody, nil
}
