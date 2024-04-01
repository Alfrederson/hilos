package api

import (
	"hilos/identity"

	"github.com/labstack/echo/v4"
)

type Session struct {
	id *identity.Identity
}

func session(c echo.Context) *Session {
	s := mustHave[Session](c, "session")
	if s == nil {
		s := &Session{}
		c.Set("session", s)
	}
	return s
}
