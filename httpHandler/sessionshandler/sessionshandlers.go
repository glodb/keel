package sessionshandler

import (
	"sync"

	"github.com/glodb/keel/settings/configmanager"

	"github.com/gorilla/sessions"
)

type session struct {
	store *sessions.CookieStore
}

var getInstance = sync.OnceValue(func() *session {
	instance := &session{}
	instance.store = sessions.NewCookieStore([]byte(configmanager.GetInstance().SessionKey))
	return instance
})

// Singleton. Returns a single object of Factory
func GetInstance() *session {
	return getInstance()
}

func (s *session) GetSession() *sessions.CookieStore {
	return s.store
}
