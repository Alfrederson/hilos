package api

import (
	"strconv"

	"github.com/labstack/echo/v4"
	"plantinha.org/m/v2/forum"
)

func ViewTopic(c echo.Context) error {
	identity := whoami(c)
	page, _ := strconv.ParseInt(c.QueryParam("p"), 32, 10)
	if page < 0 {
		page = 0
	}

	var nextPage int64
	var prevPage int64

	if page > 0 {
		prevPage = page - 1
	}

	topic, err := forum.ReadTopic(c.Param("topic_id"), page)
	if err != nil {
		return c.HTML(400, err.Error())
	}

	if (page+1)*10 < int64(topic.ReplyCount) {
		nextPage = page + 1
	}

	return c.HTML(200, RenderTemplate(
		"thread",
		R{"Topic": topic,
			"Identity": identity,
			"PrevPage": prevPage,
			"Page":     page,
			"NextPage": nextPage,
		},
	))
}
