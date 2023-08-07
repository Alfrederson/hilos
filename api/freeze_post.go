package api

import (
	"hilos/forum"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// TODO: duas funções copi-coladas.
func FreezePost(c echo.Context) error {
	post_id := c.Param("post_id")
	identity := whoami(c)

	original, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, "no post "+post_id)
	}
	if identity.Powers != 95 {
		return c.String(http.StatusForbidden, "you can't freeze posts")
	}
	log.Printf("%s freezing %s's post", identity.Name, original.Creator)

	original.Frozen = true

	if err := forum.RewritePost(post_id, original); err != nil {
		return c.String(http.StatusInternalServerError, "the forum dun gufd")
	}
	return c.HTML(http.StatusAccepted, RenderTemplate(
		"partials/post", R{
			"Post":     original,
			"Identity": identity,
		},
	))
}

func UnfreezePost(c echo.Context) error {
	post_id := c.Param("post_id")
	identity := whoami(c)

	original, err := forum.ReadPost(post_id)
	if err != nil {
		return c.String(http.StatusBadRequest, "no post "+post_id)
	}
	if identity.Powers != 95 {
		return c.String(http.StatusForbidden, "you can't freeze posts")
	}
	log.Printf("%s unfreezing %s's post", identity.Name, original.Creator)

	original.Frozen = false

	if err := forum.RewritePost(post_id, original); err != nil {
		return c.String(http.StatusInternalServerError, "the forum dun gufd")
	}
	return c.HTML(http.StatusAccepted, RenderTemplate(
		"partials/post", R{
			"Post":     original,
			"Identity": identity,
		},
	))
}
