package api

import (
	"hilos/forum"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

func Cop_ViewReports(c echo.Context) error {
	s := session(c)
	var nextPage int64 = 0
	var prevPage int64 = 0
	page := paramInt(c, "p", 0)
	// WTF: trocar isso por um middleware de paging?
	if page < 0 {
		page = 0
	}
	if page > 0 {
		prevPage = page - 1
	}
	processed := paramBool(c, "processed", false)
	reports, err := forum.GetReports(forum.GetReportOptions{
		Processed: processed,
		Page:      int(page),
		PerPage:   5,
	})
	if err != nil {
		log.Println("error getting reports:", err)
		return c.String(http.StatusInternalServerError, "the forum dun guf'd")
	}
	if len(reports) == 5 {
		nextPage = page + 1
	}
	return c.Render(200,
		"cop/reports",
		R{
			"PrevPage":  prevPage,
			"Page":      page,
			"Processed": processed,
			"NextPage":  nextPage,
			"Identity":  s.id,
			"Reports":   reports,
		},
	)
}
