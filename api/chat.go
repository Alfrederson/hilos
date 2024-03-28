package api

import (
	"github.com/labstack/echo/v4"
)

func Chat(c echo.Context) error {
	identity := whoami(c)
	return c.Render(200, "chat", R{
		"Identity": identity,
	})
}
