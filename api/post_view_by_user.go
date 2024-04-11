package api

import (
	"hilos/forum"

	"github.com/labstack/echo/v4"
)

/*
	topic, err := forum.ReadTopic(c.Param("topic_id"), page)
	if err != nil {
		return c.HTML(400, err.Error())
	}
	if (page+1)*10 < int64(topic.ReplyCount) {
		nextPage = page + 1
	}
	return c.Render(200,
		"thread",
		R{"Topic": topic,
			"Title":       topic.Subject,
			"Description": crop(topic.Content),
			"Identity":    s.id,
			"PrevPage":    prevPage,
			"Page":        page,
			"NextPage":    nextPage,
			"Elapsed":     time.Since(start),
		})
*/

func ViewUserPosts(c echo.Context) error {
	s := session(c)
	var nextPage int64 = 0
	var prevPage int64 = 0
	page := requestedPage(c)
	if page < 0 {
		page = 0
	}
	topicList, err := forum.ReadPostsByUser(c.Param("user_id"), page, 5)
	if len(topicList) == 5 {
		nextPage = page + 1
	}
	if page > 0 {
		nextPage = page - 1
	}
	if err != nil {
		return c.String(400, err.Error())
	}
	return c.Render(200,
		"user_profile",
		R{
			"Topics":   topicList,
			"Identity": s.id,
			"Page":     page,
			"NextPage": nextPage,
			"PrevPage": prevPage,
		},
	)
}
