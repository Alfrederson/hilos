package api

import (
	"log"
	"time"

	"github.com/Alfrederson/hilos/forum"

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
		return Error(c, err.Error())
	}
	return RenderPartial(c, "forms/flag_post", R{"Post": resultado})
}

func FlagPost(c echo.Context) error {
	s := session(c)
	post_id := c.Param("post_id")
	type Report struct {
		Message string `json:"message" form:"message"`
	}
	report := Report{}
	if err := c.Bind(&report); err != nil {
		log.Println("erro parseando a requisição:", err)
		return Error(c, err.Error())
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
		return Error(c, err.Error())
	}
	return Success(c, "post reported, sir")
}
