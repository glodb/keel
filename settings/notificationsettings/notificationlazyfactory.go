package notificationsettings

import (
	"errors"
	"sync"

	"github.com/glodb/keel/settings/configmanager"

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

type notifications struct {
	email          EmailSender
	latestFcm      FCMSender
	whatsappsender WhatsappSender
	activitysender ActivitySender
	// iosvoipsender  IOSVOIPSender
	semaphore *semaphore.Weighted
}

var getInstance = sync.OnceValue(func() *notifications {
	instance := &notifications{}
	instance.semaphore = semaphore.NewWeighted(int64(configmanager.GetInstance().NotificationSender.MaxConnections))
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *notifications {
	return getInstance()
}

func (c *notifications) CreateController(notificationType int) (NotificationSender, error) {

	switch notificationType {
	case SENDEMAIL:
		if !c.email.IsInitialized() {
			c.email = EmailSender{}
			c.email.Init(c.semaphore)
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
