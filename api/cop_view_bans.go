package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_ViewBans(c echo.Context) error {
	identity := whoami(c)
	if identity.Powers != 95 {
		return c.Redirect(http.StatusUnavailableForLegalReasons, "/")
	}

	return c.HTML(http.StatusAccepted, RenderTemplate(
		"cop/bans", R{
			"Bans":     "nenhum bans",
			"Identity": identity,
		},
	))
}
