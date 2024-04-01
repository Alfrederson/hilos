package api

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_ViewBans(c echo.Context) error {
	s := session(c)
	return c.HTML(http.StatusAccepted, RenderTemplate(
		"cop/bans", R{
			"Bans":     "nenhum bans",
			"Identity": s.id,
		},
	))
}
