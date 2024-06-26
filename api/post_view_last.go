package api

import (
	"github.com/Alfrederson/hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewLastPost(c echo.Context) error {
	if len(forum.Status().LastPosts) == 0 {
		return c.NoContent(200)
	}
	return RenderPartial(c, "partials/last_posts", R{"LastPosts": forum.Status().LastPosts})
}
