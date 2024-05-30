package api

import (
	"github.com/Alfrederson/hilos/forum"

	"github.com/labstack/echo/v4"
)

func Index(c echo.Context) error {
	s := session(c)
	var prevPage int64 = 0
	var nextPage int64 = 0
	page := requestedPage(c)
	topicList := forum.GetRootTopics(int(page), 5)
	if page < 0 {
		page = 0
	}
	// a gente não tem como saber se tem mais coisa...
	if len(topicList) == 5 {
		nextPage = page + 1
	}
	if page > 0 {
		prevPage = page - 1
	}
	return c.Render(200, "index", R{
		"Identity":   s.id,
		"TotalPosts": forum.Status().TotalPosts,
		"LastPosts":  forum.Status().LastPosts,
		"Topics":     topicList,
		"Page":       page,
		"PrevPage":   prevPage,
		"NextPage":   nextPage,
	})
}

func Welcome(c echo.Context) error {
	identity := whoami(c)
	return c.Render(200, "welcome", R{
		"Identity": identity,
	})
}
