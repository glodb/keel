package jsonmodels

// Intent represents what the user is asking about
type Intent string

const (
	IntentFAQ           Intent = "faq"
	IntentBookingStatus Intent = "booking_status"
	IntentCancellation  Intent = "cancellation"
	IntentUnknown       Intent = "unknown"
)

// offTopicReplies is the canned response returned when a message is not
// related to the Playpal platform. Keyed by locale ("en", "es", "ur").
var offTopicReplies = map[string]string{
	"en": "Sorry, but this chat is only to discuss about Playpal.",
	"es": "Lo siento, pero este chat es solo para hablar sobre Playpal.",
	"ur": "معذرت، یہ چیٹ صرف Playpal کے بارے میں بات کرنے کے لیے ہے۔",
}

// GetOffTopicReply returns the localised canned off-topic reply.
// Falls back to English for unknown locales.
func GetOffTopicReply(locale string) string {
	if r, ok := offTopicReplies[locale]; ok {
		return r
	}
	return offTopicReplies["en"]
}

type AgentChatRequest struct {
	UserID  string    `json:"user_id"`
	Message string    `json:"message" validate:"required"`
	Locale  string    `json:"locale" validate:"required"`
	History []Message `json:"history,omitempty"`
}

// Message is a single turn in the conversation
type Message struct {
	Role    string `json:"role"` // "user" | "model"
	Content string `json:"content"`
}

// ChatResponse is returned to the React Native app
type ChatResponse struct {
	Reply      string `json:"reply"`
	Escalate   bool   `json:"escalate"`   // true = forward to human inbox
	Confidence string `json:"confidence"` // "high" | "low"
}

// UserContext holds all fetched data relevant to the user's query
type UserContext struct {
	Intent  Intent
	Booking *BucketDataForVenue
	// Booking *BookingData
}
