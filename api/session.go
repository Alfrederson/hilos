package api

import (
	"github.com/Alfrederson/hilos/identity"

	"github.com/labstack/echo/v4"
)

type Session struct {
	id *identity.Identity
}

func session(c echo.Context) *Session {
	s := mustHave[Session](c, "session")
	// a gente torce para não ser nil, mas acho que tinha que ativar um erro aqui
	return s
}

// middleware que tira o cookie de dentro do negócio lá
// e põe a sessão dentro do c com a identidade
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
