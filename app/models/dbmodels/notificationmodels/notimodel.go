package notificationmodels

import "github.com/glodb/keel/app/models/dbmodels/keelmodels"

type NotiModels struct {
	Body              int      `json:"body"`
	Title             int      `json:"title"`
	OriginalText      string   `json:"originalText"`
	OriginalTitle     string   `json:"originalTitle"`
	ShowOriginalBody  bool     `json:"showOriginalBody"`
	ShowOriginalTitle bool     `json:"showOriginalTitle"`
	Destination       []string `json:"destination"`
	BodyData          []any    `json:"bodyData"`
	TitleData         []any    `json:"titleData"`
	Severity          int      `json:"severity"`
	Image             string   `json:"image"`
	Value             float64  `json:"value"`
	Trend             string   `json:"trend"`
	Date              int      `json:"date"`
	DeepLink          string   `json:"deepLink"`
}

type NotiReponseModels struct {
	Subject     string             `json:"subject"`
	Body        string             `json:"body"`
	Title       string             `json:"title"`
	Destination []NotificationData `json:"destination"`
	Severity    int                `json:"severity"`
	Image       string             `json:"image"`
	Value       float64            `json:"value"`
	Trend       string             `json:"trend"`
	Date        int64              `json:"date"`
}

type NotificationData struct {
	Email          string                   `json:"email,omitempty"`
	FcmKey         string                   `json:"fcmKey,omitempty"`
	IosActivityKey string                   `json:"iosActivityKey,omitempty"`
	CallLink       string                   `json:"callLink,omitempty"`
	CallType       string                   `json:"callType,omitempty"`
	CallId         string                   `json:"callId,omitempty"`
	IsVoip         bool                     `json:"isVoip,omitempty"`
	CallerName     string                   `json:"callerName,omitempty"`
	CallerImage    string                   `json:"callerImage,omitempty"`
	Phone          string                   `json:"phone,omitempty"`
	Body           string                   `json:"body"`
	Title          string                   `json:"title"`
	TitleCode      int                      `json:"titleCode,omitempty"` // Response code for title translation
	BodyCode       int                      `json:"bodyCode,omitempty"`  // Response code for body translation
	Image          string                   `json:"image"`
	DeepLink       string                   `json:"deepLink"`
	ChatData       *keelmodels.ChatUserData `json:"chatUserData,omitempty"` // Custom data payload (e.g., full chat message object)
}

type FCMMessage struct {
	Message Message `json:"message"`
}

type Message struct {
	Token        string                 `json:"token,omitempty"`
	Notification Notification           `json:"notification,omitempty"`
	Android      AndroidConfig          `json:"android,omitempty"`
	Apns         ApnsConfig             `json:"apns,omitempty"`
	Webpush      WebpushConfig          `json:"webpush,omitempty"`
	Data         map[string]interface{} `json:"data,omitempty"`
}

type AndroidConfig struct {
	Priority     string               `json:"priority,omitempty"`
	CollapseKey  string               `json:"collapse_key,omitempty"`
	FCMOptions   *AndroidFCMOptions   `json:"fcm_options,omitempty"`
	Notification *AndroidNotification `json:"notification,omitempty"`
}

type AndroidNotification struct {
	ChannelID             string `json:"channel_id,omitempty"`
	Visibility            string `json:"visibility,omitempty"`
	Title                 string `json:"title,omitempty"`
	Body                  string `json:"body,omitempty"`
	ImageURL              string `json:"image,omitempty"`
	Sound                 string `json:"sound,omitempty"`
	Sticky                bool   `json:"sticky,omitempty"`
	DefaultVibrateTimings bool   `json:"default_vibrate_timings,omitempty"`
	DefaultLightSettings  bool   `json:"default_light_settings,omitempty"`
}

type AndroidFCMOptions struct{}

type Notification struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Image string `json:"image,omitempty"`
}

type ApnsConfig struct {
	Payload           *ApnsPayload      `json:"payload,omitempty"`
	Headers           map[string]string `json:"headers,omitempty"`
	FCMOptions        *ApnsFCMOptions   `json:"fcm_options,omitempty"`
	LiveActivityToken string            `json:"live_activity_token,omitempty"`
}

type ApnsFCMOptions struct {
	ImageURL string `json:"image,omitempty"`
}

type ApnsPayload struct {
	Aps *Aps `json:"aps,omitempty"`
}

type Aps struct {
	Timestamp        int64         `json:"timestamp"`
	Event            string        `json:"event,omitempty"`
	ContentState     *ContentState `json:"content-state,omitempty"`
	Alert            *Alert        `json:"alert,omitempty"`
	ContentAvailable bool          `json:"content_available,omitempty"`
	MutableContent   bool          `json:"mutable_content,omitempty"`
	Sound            string        `json:"sound,omitempty"`
}

type ContentState struct {
	Value float64 `json:"value,omitempty"`
	Date  int     `json:"date,omitempty"`
	Trend string  `json:"trend,omitempty"`
}

type Alert struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Sound string `json:"sound,omitempty"`
}

type Title struct {
	LocKey  string   `json:"loc-key,omitempty"`
	LocArgs []string `json:"loc-args,omitempty"`
}

type Body struct {
	LocKey  string   `json:"loc-key,omitempty"`
	LocArgs []string `json:"loc-args,omitempty"`
}

type WebpushConfig struct {
	Headers      map[string]string    `json:"headers,omitempty"`
	Notification *WebpushNotification `json:"notification,omitempty"`
}

type WebpushNotification struct {
	Title string `json:"title,omitempty"`
	Body  string `json:"body,omitempty"`
	Icon  string `json:"icon,omitempty"`
}
