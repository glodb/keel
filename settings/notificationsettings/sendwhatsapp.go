package notificationsettings

import (
	"context"

	"github.com/glodb/keel/app/models/notificationmodels"

	"github.com/twilio/twilio-go"
	"golang.org/x/sync/semaphore"
)

type WhatsappSender struct {
	sendWhatsapp bool
	initialised  bool
	twilioClinet *twilio.RestClient
	semaphore    *semaphore.Weighted
}

func (u *WhatsappSender) Init(semaphore *semaphore.Weighted) (bool, error) {
	u.semaphore = semaphore

	if !u.sendWhatsapp {
		return false, nil
	}
	//TODO: Write initialisation here
	u.initialised = true
	return true, nil
}

func (u *WhatsappSender) Send(notifications []notificationmodels.NotiResponseModels) error {

	go func(u *WhatsappSender) {
		u.semaphore.Acquire(context.Background(), 1)
		defer u.semaphore.Release(1)

	}(u)
	return nil
}

func (u *WhatsappSender) MultiCastMessage(notiReponseModels notificationmodels.NotiResponseModels) error {
	u.semaphore.Acquire(context.Background(), 1)
	defer u.semaphore.Release(1)

	return nil
}

func (u *WhatsappSender) Enable() error {
	u.initialised = true
	u.sendWhatsapp = true
	return nil
}

func (u *WhatsappSender) IsInitialized() bool {
	return u.initialised
}
