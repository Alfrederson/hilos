package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewByUserId(c echo.Context) error {
	identity := whoami(c)
	topicList, err := forum.ReadUserPosts(c.Param("user_id"))
	if err != nil {
		return c.String(400, err.Error())
	}
	return c.HTML(200, RenderTemplate(
		"index",
		R{"Topics": topicList,
			"Identity": identity,
		},
	))
}
