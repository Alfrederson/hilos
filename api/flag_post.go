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
	identity := whoami(c)
	post_id := c.Param("post_id")
	type Report struct {
		Message string `json:"message" form:"message"`
	}
	report := Report{}
	if err := c.Bind(&report); err != nil {
		log.Println(err)
		return c.String(http.StatusBadRequest, "ya dun guf'd")
	}
	err := forum.FlagPost(
		post_id,
		&forum.Report{
			PostID:    post_id,
			Message:   report.Message,
			IP:        identity.IP,
			CreatorID: identity.Id,
			Time:      time.Now(),
		},
	)
	log.Println("reportado: ", post_id)
	return c.HTML(200, RenderTemplate(
		"partials/reported", R{
			"Error": err,
		},
	))
}
