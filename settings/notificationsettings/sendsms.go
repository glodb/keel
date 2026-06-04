package notificationsettings

import (
	"context"

	"github.com/glodb/keel/app/models/dbmodels/notificationmodels"

	"golang.org/x/sync/semaphore"
)

type SMSSender struct {
	sendSms     bool
	initialised bool
	semaphore   *semaphore.Weighted
}

func (u *SMSSender) Init(semaphore *semaphore.Weighted) (bool, error) {
	u.semaphore = semaphore
	if !u.sendSms {
		return false, nil
	}

	u.initialised = true
	return true, nil
}

func (u *SMSSender) Send(notifications []notificationmodels.NotiReponseModels) error {
	u.semaphore.Acquire(context.Background(), 1)
	defer u.semaphore.Release(1)

	if u.IsInitialized() {

		for _, notification := range notifications {

			var destinationStrings []string
			for _, dest := range notification.Destination {
				destinationStrings = append(destinationStrings, dest.Phone)
			}

			// url := fmt.Sprintf(configmanager.GetInstance().MoraSms.Url,
			// 	configmanager.GetInstance().MoraSms.ApiKey,
			// 	configmanager.GetInstance().MoraSms.UserName,
			// 	configmanager.GetInstance().MoraSms.Sender,
			// 	notification.Body,
			// 	strings.Join(destinationStrings, ","))

			// res, err := http.Get(url)
			// if err != nil {
			// 	return err
			// }
			// if res.StatusCode != http.StatusOK {
			// 	body, _ := io.ReadAll(res.Body)
			// 	return errors.New(string(body))
			// }
		}
	}

	return nil
}

func (u *SMSSender) MultiCastMessage(notiReponseModels notificationmodels.NotiReponseModels) error {
	u.semaphore.Acquire(context.Background(), 1)
	defer u.semaphore.Release(1)
	return nil
}

func (u *SMSSender) Enable() error {
	u.initialised = true
	u.sendSms = true
	return nil
}

func (u *SMSSender) IsInitialized() bool {
	return u.initialised
}
