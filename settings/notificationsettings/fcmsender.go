package notificationsettings

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"

	"github.com/glodb/keel/models/notificationmodels"
	"github.com/glodb/keel/settings/configmanager"
	"github.com/glodb/keel/settings/logger"
	"github.com/glodb/keel/settings/utils"

	"github.com/bytedance/sonic"

	"golang.org/x/oauth2/google"
	"golang.org/x/sync/semaphore"
)

type FCMSender struct {
	fcmNotificationSender bool
	initialised           bool
	creds                 *google.Credentials
	semaphore             *semaphore.Weighted
}

func (u *FCMSender) Init(semaphore *semaphore.Weighted) (bool, error) {
	u.semaphore = semaphore

	var credentialsJSON []byte
	var err error

	// Try to load from environment variable first (for local development)
	if envCredentials := configmanager.GetInstance().FirebaseCredentialsJson; envCredentials != "" {
		logger.Log().Info("Loading Firebase credentials from environment variable")

		// Try to decode from base64
		decoded, err := base64.StdEncoding.DecodeString(envCredentials)
		if err != nil {
			// If decode fails, assume it's already plain JSON
			logger.Log().Info("Using Firebase credentials as plain JSON from environment", logger.StringField("credentials", envCredentials))
			credentialsJSON = []byte(envCredentials)
		} else {
			logger.Log().Info("Decoded Firebase credentials from base64 environment variable")
			credentialsJSON = decoded
		}
	} else {
		// Fall back to reading from file (for production with mounted secret)
		logger.Log().Info("Loading Firebase credentials from file")
		serviceAccountKeyPath := "config/credentials/" + configmanager.GetInstance().FirebaseCredentialsFileName
		credentialsJSON = utils.GetInstance().ReadKeyFile(serviceAccountKeyPath)
	}

	// Create credentials from JSON
	creds, err := google.CredentialsFromJSON(context.Background(), credentialsJSON, "https://www.googleapis.com/auth/firebase.messaging")
	if err != nil {
		logger.Log().Error("Error loading credentials", logger.ErrorField("error", err))
		return false, err
	}

	u.creds = creds
	u.initialised = true

	if !u.fcmNotificationSender {
		return false, nil
	}

	logger.Log().Info("Firebase credentials loaded successfully")
	return true, nil
}

func (u *FCMSender) SendMesageToDestinations(notification notificationmodels.NotiResponseModels) {
	for _, destination := range notification.Destination {

		go func(destination notificationmodels.NotificationData) {

			data := map[string]interface{}{
				"deepLink": destination.DeepLink,
			}

			// Add title and body codes for translation (convert to string for FCM)
			if destination.TitleCode > 0 {
				data["titleCode"] = fmt.Sprintf("%d", destination.TitleCode)
			}
			if destination.BodyCode > 0 {
				data["bodyCode"] = fmt.Sprintf("%d", destination.BodyCode)
			}

			if destination.IsVoip {
				data["content-available"] = "1"
				data["url"] = destination.CallLink
				data["callType"] = destination.CallType
				data["conversationId"] = destination.CallId
				data["callUUID"] = destination.CallId
				data["displayName"] = destination.CallerName
				data["callerImage"] = destination.CallerImage
			}

			u.semaphore.Acquire(context.Background(), 1)
			defer u.semaphore.Release(1)

			headers := map[string]string{
				"apns-push-type": "alert",
				"apns-priority":  "10",
			}

			// Skip if NO_KEY
			if destination.FcmKey == "NO_KEY" || destination.FcmKey == "" {
				logger.Log().Warn("Skipping FCM - invalid token", logger.StringField("token", destination.FcmKey))
				return
			}

			// Log the FCM token being used
			tokenPreview := destination.FcmKey
			if len(tokenPreview) > 20 {
				tokenPreview = tokenPreview[:20] + "..."
			}
			logger.Log().Info("Sending FCM notification",
				logger.StringField("token", tokenPreview),
				logger.StringField("title", destination.Title))

			message := notificationmodels.FCMMessage{
				Message: notificationmodels.Message{
					Token: destination.FcmKey, // Replace with actual device token
					Notification: notificationmodels.Notification{
						Title: destination.Title,
						Body:  destination.Body,
						Image: destination.Image,
					},
					Android: notificationmodels.AndroidConfig{
						Priority:    "high",
						CollapseKey: "+",
						FCMOptions:  &notificationmodels.AndroidFCMOptions{},
						Notification: &notificationmodels.AndroidNotification{
							ChannelID:             "mainchannel",
							Visibility:            "public",
							Title:                 destination.Title,
							Body:                  destination.Body,
							ImageURL:              destination.Image,
							Sound:                 "default",
							Sticky:                true,
							DefaultVibrateTimings: true,
							DefaultLightSettings:  true,
						},
					},
					Apns: notificationmodels.ApnsConfig{
						Payload: &notificationmodels.ApnsPayload{
							Aps: &notificationmodels.Aps{
								ContentAvailable: true,
								MutableContent:   false,
								Sound:            "default",
							},
						},
						Headers: headers,
						FCMOptions: &notificationmodels.ApnsFCMOptions{
							ImageURL: destination.Image,
						},
					},
					Webpush: notificationmodels.WebpushConfig{
						Headers: map[string]string{
							"Urgency": "high",
						},
						Notification: &notificationmodels.WebpushNotification{
							Title: destination.Title,
							Body:  destination.Body,
							Icon:  destination.Image,
						},
					},
					Data: data, // Use the data map we created above
				},
			}

			jsonMessage, err := sonic.Marshal(message)
			if err != nil {
				logger.Log().Error("Failed to marshal message", logger.ErrorField("error", err))
				return
			}

			// req, err := http.NewRequest("POST", "https://fcm.googleapis.com/v1/projects/YOUR_PROJECT_ID/messages:send", bytes.NewBuffer(jsonMessage))
			// if err != nil {
			// 	fmt.Printf("Failed to create request: %v\n", err)
			// 	return
			// }

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
			defer resp.Body.Close()

			// Read and print the response
			body, _ := io.ReadAll(resp.Body)

			if resp.StatusCode >= 400 {
				logger.Log().Error("FCM Error Response",
					logger.IntField("status", resp.StatusCode),
					logger.StringField("token", tokenPreview),
					logger.StringField("response", string(body)))
			} else {
				logger.Log().Info("FCM Success",
					logger.IntField("status", resp.StatusCode),
					logger.StringField("token", tokenPreview))
			}
		}(destination)
	}
}

func (u *FCMSender) Send(notifications []notificationmodels.NotiResponseModels) error {

	for _, notification := range notifications {
		u.SendMesageToDestinations(notification)
	}

	return nil
}

func (u *FCMSender) MultiCastMessage(notiReponseModels notificationmodels.NotiResponseModels) error {
	u.Send([]notificationmodels.NotiResponseModels{notiReponseModels})
	return nil

}
func (u *FCMSender) Enable() error {

	u.fcmNotificationSender = true
	u.initialised = true
	return nil
}

func (u *FCMSender) IsInitialized() bool {
	return u.initialised
}
