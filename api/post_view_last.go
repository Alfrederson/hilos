package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewLastPost(c echo.Context) error {
	if len(forum.Status().LastPosts) == 0 {
		return c.NoContent(200)
	}
	return c.HTML(200,
		RenderTemplate("partials/last_posts",
			R{
				"LastPosts": forum.Status().LastPosts,
			}),
	)
}
