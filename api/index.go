package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func Index(c echo.Context) error {
	identity := whoami(c)
	topicList := forum.GetTopics(0, 100)
	return c.HTML(200,
		RenderTemplate(
			"index",
			R{
				"Identity": identity,
				"LastPost": forum.LastPost,
				"Topics":   topicList,
			},
		),
	)
}
