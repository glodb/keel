package notificationsettings

import (
	"github.com/glodb/keel/app/models/dbmodels/notificationmodels"

	"golang.org/x/sync/semaphore"
)

type NotificationSender interface {
	Enable() error
	Send(notifications []notificationmodels.NotiReponseModels) error
	MultiCastMessage(notiReponseModels notificationmodels.NotiReponseModels) error
	Init(semaphore *semaphore.Weighted) (bool, error)
	IsInitialized() bool
}
