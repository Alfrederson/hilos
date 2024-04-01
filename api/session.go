package api

import (
	"hilos/identity"
	"log"

	"github.com/labstack/echo/v4"
)

type Session struct {
	id *identity.Identity
}

func session(c echo.Context) *Session {
	s := mustHave[Session](c, "session")
	if s == nil {
		log.Println("botando uma sessão...")
		s := &Session{}
		c.Set("session", s)
	}
	return s
}

func sessionStarter(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		c.Set("session", &Session{
			id: whoami(c),
		})
		// aqui a gente vai carregar algumas coisas do arquivo,
		// tipo mensagens privadas e tal. eu acho.

		return next(c)
	}
}
