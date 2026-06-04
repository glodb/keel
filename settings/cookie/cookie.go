package cookie

import (
	"sync"

	"github.com/glodb/keel/settings/configmanager"

	"github.com/gorilla/securecookie"
)

var getInstance = sync.OnceValue(func() *cookie {
	instance := &cookie{}
	instance.secureCookie = securecookie.New([]byte(configmanager.GetInstance().SecureCookieHash), []byte(configmanager.GetInstance().SecureCookieBlock))
	return instance
})

type cookie struct {
	secureCookie *securecookie.SecureCookie
}

func GetInstance() *cookie {
	return getInstance()
}

func (c *cookie) GetCookie() *securecookie.SecureCookie {
	return c.secureCookie
}
