package api

import (
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Alfrederson/hilos/forum"

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
	post.Content = strings.TrimSpace(post.Content)
	length := len(post.Content)
	if length < 3 {
		return Error(c, "content is too short (must be at least 3 bytes)")
	}
	if length > 512 {
		return Error(c, "content is too long (can be at most 512 bytes)")
	}
	topic, err := forum.GetTopic(topic_id)
	if err != nil {
		return Error(c, "topic doesn't exist (it probably was deleted)")
	}
	if topic.Frozen && !s.id.CanMod() {
		return Error(c, "can't reply to frozen topic")
	}
	id, err := forum.WritePost(topic, post)
	if err != nil {
		log.Println("couldn't reply to topic ", topic_id, ":", err)
		return Error(c, err.Error())
	}
	post.Id = id
	return RenderPartial(c, "partials/post", R{
		"NewPost":  true,
		"Identity": s.id,
		"Post":     post,
	})
}
