package api

import (
	"log"

	"github.com/Alfrederson/hilos/forum"

	"github.com/labstack/echo/v4"
)

func Cop_DismissReport(c echo.Context) error {
	s := session(c)
	report_id := c.Param("report_id")
	report, err := forum.GetReport(report_id)
	if err != nil {
		log.Println("error dismissing report", err)
		return Error(c, err.Error())
	}
	err = forum.DismissReport(report)
	if err != nil {
		log.Println("error dismissing report", err)
		return Error(c, err.Error())
	}
	log.Println(s, " encerrou o report ", report_id)
	return Success(c, "report dismissed, sir")
}
