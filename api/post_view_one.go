package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

func ViewSinglePost(c echo.Context) error {
	post_id := c.Param("post_id")
	s := session(c)
	resultado, err := forum.ReadPost(post_id)
	if err != nil {
		return Error(c, err.Error())
	}
	return RenderPartial(c, "partials/post", R{
		"Identity": s.id,
		"Post":     resultado,
	})
}
