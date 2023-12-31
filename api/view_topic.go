package api

import (
	"hilos/forum"
	"time"

	"github.com/labstack/echo/v4"
)

func ViewTopic(c echo.Context) error {
	start := time.Now()
	identity := whoami(c)
	var nextPage int64
	var prevPage int64
	page := requestedPage(c)

	topic, err := forum.ReadTopic(c.Param("topic_id"), page)
	if err != nil {
		return c.HTML(400, err.Error())
	}

	if (page+1)*10 < int64(topic.ReplyCount) {
		nextPage = page + 1
	}

	return c.Render(200,
		"thread",
		R{"Topic": topic,
			"Title":       topic.Subject,
			"Description": crop(topic.Content),
			"Identity":    identity,
			"PrevPage":    prevPage,
			"Page":        page,
			"NextPage":    nextPage,
			"Elapsed":     time.Since(start),
		})
}
