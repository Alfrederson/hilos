package api

import (
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_ProcessReport(c echo.Context) error {
	i := whoami(c)
	if i.Powers != 95 {
		return c.String(http.StatusForbidden, "not allowed sir")
	}
	report_id := c.Param("report_id")
	log.Println(i.Name, " avaliou o report ", report_id)
	return c.String(http.StatusOK, "report processed, sir")
}
