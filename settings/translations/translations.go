package translations

import (
	"fmt"
	"sync"
)

const (
	ENGLISH = "en"
	ARABIC  = "ar"
)

const (
	SENDS_A_MESSAGE = 1002
)

type Translations struct {
	translations map[string]map[int]string
}

var getInstance = sync.OnceValue(func() *Translations {
	instance := &Translations{}
	instance.InitResponses()
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *Translations {
	return getInstance()
}

// InitResponses function just initialise the response headers to be sent
func (u *Translations) InitResponses() {
	u.translations = make(map[string]map[int]string)

	// Englisg translations
	u.translations[ENGLISH] = map[int]string{
		SENDS_A_MESSAGE: "%s Sends a message",
	}

	// Arabic translations
	u.translations[ARABIC] = map[int]string{
		SENDS_A_MESSAGE: "%s أرسل رسالة",
	}

	// Add more languages here
}

// GetResponse returns the message for the particular response code
func (u *Translations) GetResponse(code int, language string, data []any) string {
	message := ""

	// Fallback to Arabic if the requested language is not available
	if _, exists := u.translations[language]; !exists {
		language = ARABIC
	}

	// Fetch the message or return a default if the code is missing
	if msg, exists := u.translations[language][code]; exists {
		message = msg
	}

	if len(data) > 0 {
		message = fmt.Sprintf(message, data...)
	}

	return message
}
