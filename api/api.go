package api

import (
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"plantinha.org/m/v2/forum"
	"plantinha.org/m/v2/identity"

	"github.com/labstack/echo/v4"
)

type R = map[string]interface{}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

func newIdentity(c echo.Context) *identity.Identity {
	i := identity.New()
	encoded, err := i.EncodeBase64()
	if err != nil {
		log.Println("cow went to the swamp")
	} else {
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded, Path: "/"})
	}
	return &i
}

func whoami(c echo.Context) *identity.Identity {
	rwt, err := c.Cookie("rwt")
	if err != nil {
		return newIdentity(c)
	}
	i, err := identity.DecodeBase64(rwt.Value)
	if err != nil {
		return newIdentity(c)
	}
	return i
}

func Start() {
	log.Println("starting hilos...")

	e := echo.New()

	t := &Template{
		templates: template.Must(template.ParseGlob("web/*.html")),
	}

	e.Renderer = t
	// index
	e.GET("/", func(c echo.Context) error {
		topicList := forum.GetTopics(0, 100)
		identity := whoami(c)
		return c.HTML(200,
			RenderTemplate(
				"index",
				R{"Topics": topicList,
					"Identity": identity,
				},
			),
		)
	})

	e.GET("/:topic_id", func(c echo.Context) error {
		topic, err := forum.ReadTopic(c.Param("topic_id"))
		if err != nil {
			return c.HTML(400, err.Error())
		}
		identity := whoami(c)
		return c.HTML(200, RenderTemplate(
			"thread",
			R{"Topic": topic,
				"Identity": identity,
			},
		))
	})

	// view all posts by a user
	e.GET("/by/:user_id", func(c echo.Context) error {
		topicList, err := forum.ReadUserPosts(c.Param("user_id"))
		if err != nil {
			return c.String(400, err.Error())
		}
		identity := whoami(c)
		return c.HTML(200, RenderTemplate(
			"index",
			R{"Topics": topicList,
				"Identity": identity,
			},
		))
	})
	// return an identity
	e.GET("/newidentity.exe", func(c echo.Context) error {
		i := identity.New()
		encoded, err := i.EncodeBase64()
		if err != nil {
			return c.JSON(http.StatusInternalServerError, R{
				"err": "could not encode identity as base64",
			})
		}
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded})

		return c.JSON(http.StatusOK, R{
			"rwt": encoded,
		})
	})

	e.GET("/whoami.exe", func(c echo.Context) error {
		user := whoami(c)
		if user == nil {
			newIdentity := identity.New()
			user = &newIdentity
			if encoded, err := user.EncodeBase64(); err != nil {
				log.Println(err)
			} else {
				c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded})
			}
		}
		return c.JSON(http.StatusOK, user)
	})

	api := e.Group("visualbasic.exe")

	// view all topics
	api.GET("/topics.docx/:page", func(c echo.Context) error {
		page, err := strconv.ParseInt(c.Param("page"), 10, 32)
		if err != nil {
			return c.HTML(400, "wtf")
		}

		return c.JSON(http.StatusOK, forum.GetTopics(int(page), 10))
	})

	// view a single topic
	api.GET("/:topic_id", func(c echo.Context) error {
		topic, err := forum.ReadTopic(c.Param("topic_id"))
		if err != nil {
			return c.JSON(http.StatusBadRequest, R{
				"err": "could not read that topic",
			})
		}
		return c.JSON(http.StatusOK, topic)
	})

	// create a topic
	// TODO: use identity to identify the identity identification bearer
	// only tokens with OMNIPOTENCE BIT can create topics.

	api.POST("/", func(c echo.Context) error {
		nova_conversa := forum.Post{}

		if err := c.Bind(&nova_conversa); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		id, err := forum.CreateTopic(nova_conversa)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.String(http.StatusCreated, id)
	})

	// read a single post
	api.GET("/post/:topic_id", func(c echo.Context) error {
		topic_id := c.Param("topic_id")

		resultado := forum.ReadPost(topic_id)

		return c.JSON(http.StatusOK, resultado)
	})

	// reply a thread
	// TODO: use identity to identify the identity identification bearer
	api.POST("/post/:topic_id", func(c echo.Context) error {
		post := forum.Post{}
		if err := c.Bind(&post); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		topic_id := c.Param("topic_id")
		identity := whoami(c)

		post.CreatorId = identity.Id
		post.Creator = identity.Name

		id, err := forum.ReplyTopic(topic_id, post)
		if err != nil {
			log.Println("couldn't reply to topic ", topic_id, ":", err)
			return c.String(http.StatusBadRequest, "could not record the message")
		}
		return c.String(http.StatusAccepted, RenderTemplate(
			"newpost", R{
				"Id":        id,
				"Subject":   post.Subject,
				"Creator":   identity.Name,
				"CreatorId": identity.Id,
				"Content":   post.Content,
			},
		))
	})

	e.Start(":3000")
}
