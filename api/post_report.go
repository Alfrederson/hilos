package api

import (
	"hilos/forum"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

// Isso funciona assim:
// - pessoa clica no botão "🚩report"
// - vê se já tem um report com o IP daquele post.
// - se não tiver, coloca ele na fila.

func FormFlagPost(c echo.Context) error {
	post_id := c.Param("post_id")
	resultado, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.HTML(200, RenderTemplate(
		"forms/flag_post",
		R{
			"Post": resultado,
		},
	))
}

func FlagPost(c echo.Context) error {
	s := session(c)
	post_id := c.Param("post_id")
	type Report struct {
		Message string `json:"message" form:"message"`
	}
	report := Report{}
	if err := c.Bind(&report); err != nil {
		log.Println(err)
		return c.HTML(http.StatusAccepted, RenderTemplate("alert/error", R{"Message": err}))
	}
	if err := forum.ReportPost(
		post_id,
		&forum.Report{
			PostID:      post_id,
			Message:     report.Message,
			IP:          s.id.IP,
			CreatorID:   s.id.Id,
			CreatorName: s.id.Name,
			Time:        time.Now(),
		},
	); err != nil {
		return c.HTML(http.StatusAccepted, RenderTemplate("alert/error", R{"Message": err}))
	}
	return c.HTML(http.StatusAccepted, RenderTemplate("alert/success", R{"Message": "post reported sir"}))
}
