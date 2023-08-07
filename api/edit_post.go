package api

import (
	"hilos/forum"
	"log"
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

func EditPost(c echo.Context) error {
	type Alteration struct {
		Subject string `json:"subject" form:"subject"`
		Content string `json:"content" form:"content"`
	}
	post_id := c.Param("post_id")
	identity := whoami(c)

	changes := Alteration{}
	if err := c.Bind(&changes); err != nil {
		log.Println(err)
		return c.String(http.StatusBadRequest, "ya dun guf'd")
	}

	original, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, "no post "+post_id)
	}
	// Tirar aquele lá de cima quando eu descobrir um jeito prático de
	// exibir os posts...
	if (original.CreatorId != identity.Id) && (identity.Powers != 95) {
		return c.String(http.StatusForbidden, "you can't edit someone else's post")
	}

	log.Printf("%s editing %s's post", identity.Name, original.Creator)
	original.Subject = changes.Subject
	original.Content = changes.Content

	if err := forum.RewritePost(post_id, original); err != nil {
		return c.String(http.StatusInternalServerError, "the forum dun gufd")
	}

	return c.HTML(http.StatusAccepted, RenderTemplate(
		"partials/post", R{
			"Post":     original,
			"Identity": identity,
		},
	))
}
