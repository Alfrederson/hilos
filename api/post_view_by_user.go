package api

import (
	"hilos/forum"
	"strconv"

	"github.com/labstack/echo/v4"
)

func ViewUserPosts(c echo.Context) error {
	s := session(c)
	fromPage, _ := strconv.ParseInt(c.QueryParam("p"), 10, 32)
	topicList, err := forum.ReadPostsByUser(c.Param("user_id"), fromPage)
	if err != nil {
		return c.String(400, err.Error())
	}
	return c.Render(200,
		"index",
		R{
			"Topics":    topicList,
			"LastPosts": forum.Status().LastPosts,
			"Identity":  s.id,
		},
	)
}
