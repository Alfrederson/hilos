package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewLastPost(c echo.Context) error {
	if forum.LastPost == nil {
		return c.NoContent(200)
	}

	return c.HTML(200, RenderTemplate("lastpost", R{"LastPost": forum.LastPost}))
}
