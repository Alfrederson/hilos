package api

import (
	"log"
	"net/http"

	"github.com/Alfrederson/hilos/forum"

	"github.com/labstack/echo/v4"
)

// Isso funciona assim:
// - pessoa clica no botão "Editar."
// - sistema checa se a pessoa é Admin ou se o id dela bate com o ID de quem fez o post.
// - se bater, retorna um formulário pra editar.
//
// - quando o formulário é enviado, envia aquele troço lá.

func FormEditPost(c echo.Context) error {
	s := session(c)
	topic_id := c.Param("post_id")
	post, err := forum.ReadPost(topic_id)
	if err != nil {
		return c.String(http.StatusBadRequest, " post no foundy")
	}
	if post.CreatorId != s.id.Id && !s.id.CanMod() {
		return c.String(http.StatusForbidden, " cant edit post, sir")
	}
	return RenderPartial(c, "forms/edit_post", R{
		"Id":      post.Id,
		"Subject": post.Subject,
		"Content": post.Content,
	})
}

func EditPost(c echo.Context) error {
	type Alteration struct {
		Subject string `json:"subject" form:"subject"`
		Content string `json:"content" form:"content"`
	}
	s := session(c)
	post_id := c.Param("post_id")
	changes := Alteration{}
	if err := c.Bind(&changes); err != nil {
		log.Println(err)
		return Error(c, err.Error())
	}
	original, err := forum.ReadPost(post_id)
	if err != nil {
		return Error(c, err.Error())
	}
	// Tirar aquele lá de cima quando eu descobrir um jeito prático de
	// exibir os posts...
	if (original.CreatorId != s.id.Id) && !s.id.CanMod() {
		return Error(c, "you can't edit someone else's post")
	}
	log.Printf("%s editing %s's post", s.id.Name, original.Creator)
	original.Subject = changes.Subject
	original.Content = changes.Content
	if err := forum.RewritePost(post_id, original); err != nil {
		return Error(c, err.Error())
	}
	return RenderPartial(c, "partials/post", R{
		"Post":     original,
		"Identity": s.id,
	})
}
