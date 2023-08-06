package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewTopic(c echo.Context) error {
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
			"Identity": identity,
			"PrevPage": prevPage,
			"Page":     page,
			"NextPage": nextPage,
		})
}
