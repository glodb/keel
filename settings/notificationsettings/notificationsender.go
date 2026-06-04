package notificationsettings

import (
	"github.com/glodb/keel/models/notificationmodels"

	"golang.org/x/sync/semaphore"
)

type NotificationSender interface {
	Enable() error
	Send(notifications []notificationmodels.NotiResponseModels) error
	MultiCastMessage(notiReponseModels notificationmodels.NotiResponseModels) error
	Init(semaphore *semaphore.Weighted) (bool, error)
	IsInitialized() bool
}
