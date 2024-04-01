package api

import (
	"hilos/forum"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TODO: duas funções copi-coladas.
func Cop_FreezePost(c echo.Context) error {
	s := session(c)
	post_id := c.Param("post_id")
	original, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusNotFound, "post "+post_id+" doesn't exist")
	}
	original.Frozen = true
	if err := forum.RewritePost(post_id, original); err != nil {
		return c.String(http.StatusInternalServerError, "the forum dun gufd")
	}
	return c.HTML(http.StatusAccepted, RenderTemplate(
		"partials/post", R{
			"Post":     original,
			"Identity": s.id,
		},
	))
}

func Cop_UnfreezePost(c echo.Context) error {
	post_id := c.Param("post_id")
	s := session(c)

	original, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, "no post "+post_id)
	}
	log.Printf("%s unfreezing %s's post", s.id.Name, original.Creator)

	original.Frozen = false

	if err := forum.RewritePost(post_id, original); err != nil {
		return c.String(http.StatusInternalServerError, "the forum dun gufd")
	}
	return c.HTML(http.StatusAccepted, RenderTemplate(
		"partials/post", R{
			"Post":     original,
			"Identity": s.id,
		},
	))
}
