package api

import (
	"hilos/forum"
	"net/http"

	"github.com/labstack/echo/v4"
)

func ViewSinglePost(c echo.Context) error {
	post_id := c.Param("post_id")
	s := session(c)
	resultado, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, err.Error())
	}
	return c.HTML(200, RenderTemplate(
		"partials/post",
		R{
			"Identity": s.id.Id,
			"Post":     resultado,
		},
	))
}
