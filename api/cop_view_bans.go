package api

import (
	"github.com/labstack/echo/v4"
)

// WTF: Por que isso manda um parcial?
// acho que ele vai baixar a lista de bans depois de entrar
// na telinha.
func Cop_ViewBans(c echo.Context) error {
	s := session(c)
	return RenderPartial(c, "cop/bans", R{
		"Identity": s.id,
		"Bans":     "nenhum bans",
	})
}
