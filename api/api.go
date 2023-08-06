package api

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"

	"hilos/forum"
	"hilos/identity"

	"github.com/labstack/echo/v4"
)

type R = map[string]interface{}

type Template struct {
	templates *template.Template
}

func (t *Template) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	return t.templates.ExecuteTemplate(w, name, data)
}

type WebConfig struct {
	Prefix string
}

var web = WebConfig{
	Prefix: "thread.docx",
}

func Start() {
	e := echo.New()

	pugger := MakePugger()
	e.Renderer = pugger

	//  ainda não sei como fazer isso funcionar, mas supostamente é pra acelerar um pouco
	//  a geração das páginas. TODO: fazer isso.

	//	e.Renderer = t
	e404 := func(c echo.Context) error {
		return c.String(404, "not found")
	}

	ç := func(c echo.Context) error {
		log.Println(c.RealIP(), " looking for ", c.Request().RequestURI)
		return c.String(200, "🤡")
	}

	// The Index
	e.GET("/favicon.ico", e404)
	e.GET("/robots.txt", func(c echo.Context) error {
		log.Println("🕷️ ", c.RealIP())
		return c.String(200, "User-agent: *\nDisallow:")
	})
	e.GET("/admin.php", ç)
	e.GET("/wp-admin/", ç)
	e.GET("/.env", ç)

	e.GET("/", Index)
	// View a thread/topic/post whatever
	e.GET("/:topic_id", ViewTopic)

	// View the last post
	e.GET("/new", ViewLastPost)

	// view all posts by a user
	e.GET("/by/:user_id", ViewByUserId)

	// creates a new identity
	e.GET("/newidentity.exe", func(c echo.Context) error {
		i := identity.New()
		encoded, err := i.EncodeBase64()
		if err != nil {
			return c.String(http.StatusInternalServerError, "could not send your new id")
		}
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded})
		return c.String(http.StatusOK, encoded)
	})

	e.GET("fakepassport.exe/:passport", func(c echo.Context) error {
		i, err := identity.DecodeBase64(c.Param("passport"))
		if err != nil {
			log.Println("fake passport: ", err)
			return c.String(http.StatusPaymentRequired, "need to pay bribe for fake passport. its very fake.")
		}
		if !i.Check() {
			return c.String(http.StatusConflict, "sorry sir this isn't accepted")
		}
		encoded, err := i.EncodeBase64()
		if err != nil {
			return c.String(http.StatusConflict, "could not encode fake passport")
		}
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded, Path: "/"})
		return c.String(http.StatusOK, fmt.Sprintf("you are now %s %s with powers %d", i.Id, i.Name, i.Powers))
	})

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

	// edit a post
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
		if (original.CreatorId != identity.Id) && (identity.Powers != 95) {
			return c.String(http.StatusForbidden, "ya dun gufd")
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

		return c.HTML(200, RenderTemplate(
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
