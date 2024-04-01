package api

import (
	"hilos/forum"
	"log"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
)

func ReplyThread(c echo.Context) error {
	topic_id := c.Param("topic_id")
	s := session(c)
	post := forum.Post{
		CreatorId: s.id.Id,
		Creator:   s.id.Name,
		IP:        s.id.IP,
		Time:      time.Now(),
	}
	if err := c.Bind(&post); err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	length := len(post.Content)
	if length < 3 {
		return c.String(http.StatusBadRequest, "content too short, sir")
	}
	if length > 512 {
		return c.String(http.StatusBadRequest, "content too long, sir")
	}
	topic, err := forum.GetTopic(topic_id)
	if err != nil {
		return c.String(http.StatusNotFound, "no such topic")
	}
	if topic.Frozen && !s.id.CanMod() {
		return c.String(http.StatusNotFound, "cannot reply frozen topic")
	}
	id, err := forum.WritePost(topic, post)
	if err != nil {
		log.Println("couldn't reply to topic ", topic_id, ":", err)
		return c.String(http.StatusBadRequest, "could not record the message")
	}
	post.Id = id
	return c.HTML(200, RenderTemplate(
		"partials/post", R{
			"NewPost":  true,
			"Identity": s.id,
			"Post":     post,
		},
	))
}
