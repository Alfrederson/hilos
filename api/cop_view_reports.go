package api

import (
	"hilos/forum"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_ViewReports(c echo.Context) error {
	identity := whoami(c)
	if identity.Powers != 95 {
		return c.Redirect(451, "/")
	}
	reports, err := forum.GetReports()
	if err != nil {
		log.Println("error getting reports:", err)
		return c.String(http.StatusInternalServerError, "the forum dun guf'd")
	}

	return c.Render(200,
		"cop/reports",
		R{
			"Identity": identity,
			"Reports":  reports,
		},
	)
}
