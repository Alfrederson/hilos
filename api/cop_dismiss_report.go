package api

import (
	"hilos/forum"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_DismissReport(c echo.Context) error {
	s := session(c)

	report_id := c.Param("report_id")
	report, err := forum.GetReport(report_id)
	if err != nil {
		log.Println("error dismissing report", err)
		return c.String(http.StatusNotFound, " report not found sir")
	}
	err = forum.DismissReport(report)
	if err != nil {
		log.Println("error dismissing report", err)
		return c.String(http.StatusBadRequest, " could not dismiss report, sir")
	}
	log.Println(s, " encerrou o report ", report_id)
	return c.String(http.StatusOK, "report dismissed, sir")
}
