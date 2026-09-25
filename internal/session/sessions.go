package session

import (
	"net/http"
	"time"

	"github.com/alexedwards/scs/v2"
)

func New(inProduction bool) *scs.SessionManager {
	manager := scs.New()

	manager.Lifetime = 24 * time.Hour
	manager.Cookie.Persist = true
	manager.Cookie.SameSite = http.SameSiteLaxMode
	manager.Cookie.Secure = inProduction

	return manager
}
