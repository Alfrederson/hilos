package api

import (
	"hilos/forum"
	"log"
	"net/http"

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
	post_id := c.Param("post_id")
	reportado, err := forum.ReadPost(post_id)
	log.Println(err)
	if err != nil {
		return c.HTML(200, "WTF")
	}
	return c.HTML(200, RenderTemplate(
		"partials/post",
		R{
			"Post": reportado,
		},
	))
}
