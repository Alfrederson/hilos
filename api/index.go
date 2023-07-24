package api

import (
	"github.com/labstack/echo/v4"
	"plantinha.org/m/v2/forum"
)

func Index(c echo.Context) error {
	identity := whoami(c)
	topicList := forum.GetTopics(0, 100)
	return c.HTML(200,
		RenderTemplate(
			"index",
			R{"Topics": topicList,
				"Identity": identity,
			},
		),
	)
}
