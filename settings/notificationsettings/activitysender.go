package notificationsettings

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/glodb/keel/app/models/notificationmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utils"

	"github.com/bytedance/sonic"
	"golang.org/x/oauth2/google"
	"golang.org/x/sync/semaphore"
)

type ActivitySender struct {
	notificationSender bool
	initialised        bool
	creds              *google.Credentials
	// client             *apns2.Client
	semaphore *semaphore.Weighted
}

func (u *ActivitySender) Init(semaphore *semaphore.Weighted) (bool, error) {
	u.semaphore = semaphore

	serviceAccountKeyPath := "config/credentials/" + configmanager.GetInstance().FirebaseCredentialsFileName

	creds, err := google.CredentialsFromJSON(context.Background(), utils.GetInstance().ReadKeyFile(serviceAccountKeyPath), "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		logger.Log().Error("Error loading credentials", logger.ErrorField("error", err))
		return false, err
	}

	u.creds = creds

	u.initialised = true

	if !u.notificationSender {
		return false, nil
	}

	return true, nil
}

func (u *ActivitySender) SendMesageToDestinations(notification notificationmodels.NotiResponseModels) {
	for _, destination := range notification.Destination {

		go func(notification notificationmodels.NotiResponseModels, destination notificationmodels.NotificationData) {

			u.semaphore.Acquire(context.Background(), 1)
			defer u.semaphore.Release(1)

			headers := map[string]string{
				"apns-priority": "10",
			}

			message := notificationmodels.FCMMessage{
				Message: notificationmodels.Message{
					Token: destination.FcmKey,
					Apns: notificationmodels.ApnsConfig{
						LiveActivityToken: destination.IosActivityKey,
						Headers:           headers,
						Payload: &notificationmodels.ApnsPayload{
							Aps: &notificationmodels.Aps{
								Timestamp: time.Now().Unix(),
								Event:     "update",
								ContentState: &notificationmodels.ContentState{
									Value: notification.Value,
									Date:  int(notification.Date),
									Trend: notification.Trend,
								},
							},
						},
					},
				},
			}

			jsonMessage, err := sonic.Marshal(message)
			if err != nil {
				logger.Log().Error("Failed to marshal message", logger.ErrorField("error", err))
				return
			}

			// Create the HTTP client
			client := utils.GetInstance().Oauth2HTTPClient(context.Background(), u.creds)

			url := fmt.Sprintf(configmanager.GetInstance().FirebaseMessageUrl, configmanager.GetInstance().FirebaseProjectId)
			req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonMessage))
			if err != nil {
				logger.Log().Error("Error creating request", logger.ErrorField("error", err))
				return
			}
			accessToken, err := u.creds.TokenSource.Token()

			if err != nil {
				logger.Log().Error("Error sending request", logger.ErrorField("error", err))
				return
			}

			req.Header.Set("Authorization", "Bearer "+accessToken.AccessToken)
			req.Header.Set("Content-Type", "application/json")

			resp, err := client.Do(req)
			if err != nil {
				logger.Log().Error("Error sending request", logger.ErrorField("error", err))
				return
			}

			if resp.StatusCode != http.StatusOK {
				logger.Log().Error("Unexpected status code", logger.IntField("status_code", resp.StatusCode), logger.StringField("fcm_key", destination.FcmKey), logger.StringField("live_activity_code", destination.IosActivityKey))
				body, _ := io.ReadAll(resp.Body)
				logger.Log().Error("Response", logger.StringField("response", string(body)))
				return
			}

			defer resp.Body.Close()

			// Read and print the response
			body, _ := io.ReadAll(resp.Body)

			logger.Log().Debug("Successful Response", logger.StringField("response", string(body)))
		}(notification, destination)
	}
}

func (u *ActivitySender) Send(notifications []notificationmodels.NotiResponseModels) error {

	for _, notification := range notifications {
		u.SendMesageToDestinations(notification)
	}
	return nil
}

func (u *ActivitySender) MultiCastMessage(notiReponseModels notificationmodels.NotiResponseModels) error {
	u.Send([]notificationmodels.NotiResponseModels{notiReponseModels})

	return nil
}

func (u *ActivitySender) Enable() error {

	u.notificationSender = true
	u.initialised = true
	return nil
}

func (u *ActivitySender) IsInitialized() bool {
	return u.initialised
}
