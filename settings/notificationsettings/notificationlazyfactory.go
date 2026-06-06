package notificationsettings

import (
	"errors"
	"sync"

	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"

	"golang.org/x/sync/semaphore"
)

const (
	SENDEMAIL = iota
	LATESTFCM
	SENDWHATSAPP
	SENDSMS
	ACTIVITYSENDER
	IOSVOIP
)

type Notifications struct {
	email          EmailSender
	latestFcm      FCMSender
	whatsappsender WhatsappSender
	activitysender ActivitySender
	// iosvoipsender  IOSVOIPSender
	semaphore *semaphore.Weighted
}

var getInstance = sync.OnceValue(func() *Notifications {
	instance := &Notifications{}
	instance.semaphore = semaphore.NewWeighted(int64(configmanager.GetInstance().NotificationSender.MaxConnections))
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *Notifications {
	return getInstance()
}

func NewNotifications(maxConnections int) *Notifications {
	instance := &Notifications{}
	instance.semaphore = semaphore.NewWeighted(int64(maxConnections))
	return instance
}

func (c *Notifications) CreateController(notificationType int) (NotificationSender, error) {

	switch notificationType {
	case SENDEMAIL:
		if !c.email.IsInitialized() {
			c.email = EmailSender{}
			if _, err := c.email.Init(c.semaphore); err != nil {
				logger.Log().Error("email sender init failed", logger.ErrorField("error", err))
				c.email = EmailSender{}
				return nil, err
			}
			c.email.Enable()
		}
		return &c.email, nil
	case LATESTFCM:
		if !c.latestFcm.IsInitialized() {
			c.latestFcm = FCMSender{}
			c.latestFcm.Init(c.semaphore)
			c.latestFcm.Enable()
		}
		return &c.latestFcm, nil
	case SENDWHATSAPP:
		if !c.whatsappsender.IsInitialized() {
			c.whatsappsender = WhatsappSender{}
			c.whatsappsender.Init(c.semaphore)
			c.whatsappsender.Enable()
		}
		return &c.whatsappsender, nil

	case ACTIVITYSENDER:
		if !c.activitysender.IsInitialized() {
			c.activitysender = ActivitySender{}
			c.activitysender.Init(c.semaphore)
			c.activitysender.Enable()
		}
		return &c.activitysender, nil
	}
	return nil, errors.New("not known notification type")
}
