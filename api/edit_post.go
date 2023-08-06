package api

import (
	"hilos/forum"
	"net/http"

	"github.com/labstack/echo/v4"
)

// Isso funciona assim:
// - pessoa clica no botão "Editar."
// - sistema checa se a pessoa é Admin ou se o id dela bate com o ID de quem fez o post.
// - se bater, retorna um formulário pra editar.
//
// - quando o formulário é enviado, envia aquele troço lá.

func FormEditPost(c echo.Context) error {
	topic_id := c.Param("post_id")

	resultado, err := forum.ReadPost(topic_id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}

	return c.HTML(200, RenderTemplate(
		"forms/edit_post",
		R{
			"Id":      resultado.Id,
			"Subject": resultado.Subject,
			"Content": resultado.Content,
		},
	))
}
