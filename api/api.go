package api

import (
	"fmt"
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
	if !i.Check() {
		return newIdentity(c)
	}
	return i
}

func Start() {
	e := echo.New()

	//  ainda não sei como fazer isso funcionar, mas supostamente é pra acelerar um pouco
	//  a geração das páginas. TODO: fazer isso.
	//	t := &Template{
	//		templates: template.Must(template.ParseGlob("web/*.html")),
	//	}

	//	e.Renderer = t
	e404 := func(c echo.Context) error {
		return c.String(404, "not found")
	}

	// The Index
	e.GET("/", func(c echo.Context) error {
		identity := whoami(c)
		topicList := forum.GetTopics(0, 100)
		return c.HTML(200,
			RenderTemplate(
				"index",
				R{"Topics": topicList,
					"Identity": identity,
				},
			),
		)
	})
	e.GET("/favicon.ico", e404)
	e.GET("/robots.txt", func(c echo.Context) error {
		log.Println("🕷️ ", c.RealIP())
		return c.String(200, "User-agent: *\nDisallow:")
	})

	// View a thread/topic/post whatever
	e.GET("/:topic_id", func(c echo.Context) error {
		identity := whoami(c)
		page, _ := strconv.ParseInt(c.QueryParam("p"), 32, 10)
		if page < 0 {
			page = 0
		}

		var nextPage int64
		var prevPage int64

		if page > 0 {
			prevPage = page - 1
		}

		topic, err := forum.ReadTopic(c.Param("topic_id"), page)
		if err != nil {
			return c.HTML(400, err.Error())
		}

		if (page+1)*10 < int64(topic.ReplyCount) {
			nextPage = page + 1
		}

		return c.HTML(200, RenderTemplate(
			"thread",
			R{"Topic": topic,
				"Identity": identity,
				"PrevPage": prevPage,
				"Page":     page,
				"NextPage": nextPage,
			},
		))
	})

	// view all posts by a user
	e.GET("/by/:user_id", func(c echo.Context) error {
		identity := whoami(c)
		topicList, err := forum.ReadUserPosts(c.Param("user_id"))
		if err != nil {
			return c.String(400, err.Error())
		}
		return c.HTML(200, RenderTemplate(
			"index",
			R{"Topics": topicList,
				"Identity": identity,
			},
		))
	})
	// creates a new identity
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
	// reuse an identity
	/*
		e.POST("fakepassport.exe", func(c echo.Context) error{
			// TODO: validar o corpo da requisição e retornar um cookie.
		})
	*/

	e.GET("/whoami.exe", func(c echo.Context) error {
		user := whoami(c)
		return c.String(http.StatusOK, fmt.Sprint(user.Id, user.Name))
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
		page, _ := strconv.ParseInt(c.QueryParam("p"), 32, 10)
		if page < 0 {
			page = 0
		}

		topic, err := forum.ReadTopic(c.Param("topic_id"), page)

		if err != nil {
			return c.JSON(http.StatusBadRequest, R{
				"err": "could not read that topic",
			})
		}
		return c.JSON(http.StatusOK, topic)
	})

	// create a topic
	// TODO: use identity to identify the identity identification bearer
	// only tokens with OMNIPOTENCE BIT (0d95) can create topics.
	api.POST("/", func(c echo.Context) error {
		identity := whoami(c)
		if identity.Powers != 95 {
			return e404(c)
		}

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

		resultado, err := forum.ReadPost(topic_id)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		return c.JSON(http.StatusOK, resultado)
	})

	api.PUT("/post/:post_id", func(c echo.Context) error {
		type Alteration struct {
			Subject string `json:"subject" form:"subject"`
			Content string `json:"content" form:"content"`
		}

		post_id := c.Param("post_id")
		identity := whoami(c)
		if identity.Powers != 95 {
			return e404(c)
		}

		changes := Alteration{}

		if err := c.Bind(&changes); err != nil {
			log.Println(err)
			return c.String(http.StatusBadRequest, "ya dun guf'd")
		}

		// TODO: limit this to people que tem plenos poderes sobre o forum (poderes 0d95) ou que
		// ainda tem créditos.
		// no momento qualquer pessoa pode editar qualquer post de qualquer outra pessoa.

		original, err := forum.ReadPost(post_id)
		if err != nil {
			return c.String(http.StatusBadRequest, "no post "+post_id)
		}

		log.Printf("%s editing %s's post", identity.Name, original.Creator)

		original.Subject = changes.Subject
		original.Content = changes.Content

		if err := forum.RewritePost(post_id, original); err != nil {
			return c.String(http.StatusInternalServerError, "the forum dun gufd")
		}

		return c.HTML(http.StatusAccepted, RenderTemplate(
			"newpost", R{
				"Id":        original.Id,
				"Subject":   original.Subject,
				"Creator":   original.Creator,
				"CreatorId": original.CreatorId,
				"Content":   original.Content,
			},
		))
	})

	// reply a thread
	api.POST("/post/:topic_id", func(c echo.Context) error {
		topic_id := c.Param("topic_id")
		identity := whoami(c)

		post := forum.Post{}
		if err := c.Bind(&post); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		length := len(post.Content)
		if length < 3 {
			return c.String(http.StatusBadRequest, "content too short, sir")
		}
		if length > 512 {
			return c.String(http.StatusBadRequest, "content too long, sir")
		}

		post.CreatorId = identity.Id
		post.Creator = identity.Name

		id, err := forum.ReplyTopic(topic_id, post)
		if err != nil {
			log.Println("couldn't reply to topic ", topic_id, ":", err)
			return c.String(http.StatusBadRequest, "could not record the message")
		}
		return c.HTML(http.StatusAccepted, RenderTemplate(
			"newpost", R{
				"Id":        id,
				"Subject":   post.Subject,
				"Creator":   identity.Name,
				"CreatorId": identity.Id,
				"Content":   post.Content,
			},
		))
	})

	if err := e.Start(":3000"); err != nil {
		panic(err)
	}
}
