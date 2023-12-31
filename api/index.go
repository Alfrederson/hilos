package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func Index(c echo.Context) error {
	identity := whoami(c)
	page := requestedPage(c)
	topicList := forum.GetTopics(int(page), 10)

	return c.Render(200, "index", R{
		"Identity":  identity,
		"LastPosts": forum.Status().LastPosts,
		"Topics":    topicList,
	})
}
