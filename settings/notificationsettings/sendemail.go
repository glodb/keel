package notificationsettings

import (
	"context"
	"crypto/tls"
	"strconv"
	"strings"

	"github.com/glodb/keel/models/notificationmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"

	"golang.org/x/sync/semaphore"
	"gopkg.in/gomail.v2"
)

type EmailSender struct {
	sendEmail   bool
	initialised bool
	semaphore   *semaphore.Weighted
	dialer      *gomail.Dialer
	sender      gomail.SendCloser
}

func (u *EmailSender) Init(semaphore *semaphore.Weighted) (bool, error) {
	u.semaphore = semaphore
	portToInt, err := strconv.Atoi(configmanager.GetInstance().Email.SMPTPort)
	logger.Log().Debug("Settings",
		logger.StringField("SMPTClient", configmanager.GetInstance().Email.SMPTClient),
		logger.StringField("SMPTPort", configmanager.GetInstance().Email.SMPTPort),
		logger.StringField("EmailAddress", configmanager.GetInstance().Email.EmailAddress),
		logger.StringField("Password", configmanager.GetInstance().Email.Password),
		logger.StringField("SMPTClient", configmanager.GetInstance().Email.SMPTClient),
		logger.BoolField("IsTLS", configmanager.GetInstance().Email.IsTLS),
		logger.BoolField("InsecureSkipVerify", configmanager.GetInstance().Email.InsecureSkipVerify),
	)
	if err != nil {
		return false, err
	}
	u.dialer = &gomail.Dialer{
		Host:     configmanager.GetInstance().Email.SMPTClient,
		Port:     portToInt,
		Username: configmanager.GetInstance().Email.EmailAddress,
		Password: configmanager.GetInstance().Email.Password,
		SSL:      configmanager.GetInstance().Email.IsTLS,
		TLSConfig: &tls.Config{
			InsecureSkipVerify: configmanager.GetInstance().Email.InsecureSkipVerify,
		},
	}

	s, err := u.dialer.Dial()
	if err != nil {
		return false, err
	}
	u.sender = s
	u.initialised = true
	return true, nil
}

func (u *EmailSender) reconnect() error {
	s, err := u.dialer.Dial()
	if err != nil {
		return err
	}
	u.sender = s
	return nil
}

func (u *EmailSender) Send(notifications []notificationmodels.NotiResponseModels) error {
	u.semaphore.Acquire(context.Background(), 1)
	defer u.semaphore.Release(1)

	for _, notification := range notifications {
		message := gomail.NewMessage()

		senderName := ""
		toEmail := notification.Destination[0].Email // one email address per notification
		if configmanager.GetInstance().Email.EmailName != "" {
			senderName = message.FormatAddress(configmanager.GetInstance().Email.EmailAddress, configmanager.GetInstance().Email.EmailName)
		} else {
			senderName = configmanager.GetInstance().Email.EmailAddress
		}
		message.SetHeader("From", senderName)
		message.SetHeader("To", toEmail)
		message.SetHeader("Subject", notification.Title)
		message.SetBody("text/html", notification.Body)

		if err := gomail.Send(u.sender, message); err != nil {
			logger.Log().Warn("error sending email; attempting reconnect", logger.ErrorField("error", err))
			if rerr := u.reconnect(); rerr != nil {
				logger.Log().Error("smtp reconnect failed", logger.ErrorField("error", rerr))
				return err
			}
			if err2 := gomail.Send(u.sender, message); err2 != nil {
				logger.Log().Error("error sending email after reconnect", logger.ErrorField("error", err2), logger.StringField("to", toEmail), logger.StringField("subject", notification.Title))
				return err2
			}
		}
	}

	return nil
}

func (u *EmailSender) MultiCastMessage(notification notificationmodels.NotiResponseModels) error {
	u.semaphore.Acquire(context.Background(), 1)
	defer u.semaphore.Release(1)

	toEmails := []string{}
	for _, destination := range notification.Destination {
		toEmails = append(toEmails, destination.Email)
	}
	message := gomail.NewMessage()
	senderName := ""
	if configmanager.GetInstance().Email.EmailName != "" {
		senderName = message.FormatAddress(configmanager.GetInstance().Email.EmailAddress, configmanager.GetInstance().Email.EmailName)
	} else {
		senderName = configmanager.GetInstance().Email.EmailAddress
	}
	message.SetHeader("From", senderName)
	message.SetHeader("To", toEmails...)
	message.SetHeader("Subject", notification.Title)
	message.SetBody("text/html", notification.Body)

	if err := gomail.Send(u.sender, message); err != nil {
		logger.Log().Warn("error sending email; attempting reconnect", logger.ErrorField("error", err))
		if rerr := u.reconnect(); rerr != nil {
			logger.Log().Error("smtp reconnect failed", logger.ErrorField("error", rerr))
			return err
		}
		if err2 := gomail.Send(u.sender, message); err2 != nil {
			logger.Log().Error("error sending email after reconnect", logger.ErrorField("error", err2), logger.StringField("to", strings.Join(toEmails, ",")), logger.StringField("subject", notification.Title))
			return err2
		}
	}

	return nil
}

func (u *EmailSender) Close() error {
	if u.sender != nil {
		return u.sender.Close()
	}
	return nil
}

func (u *EmailSender) Enable() error {
	u.initialised = true
	u.sendEmail = true
	return nil
}

func (u *EmailSender) IsInitialized() bool {
	return u.initialised
}
