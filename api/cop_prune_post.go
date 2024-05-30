package api

import (
	"github.com/Alfrederson/hilos/forum"

	"github.com/labstack/echo/v4"
)

func Cop_PrunePost(c echo.Context) error {
	post_id := c.Param("post_id")
	err := forum.PrunePost(post_id)
	if err != nil {
		return Error(c, "oops: "+err.Error())
	}
	return Success(c, "post has been pruned. children posts still exist, but will be removed over time.")
}
