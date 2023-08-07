package api

import (
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"

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

	// view a single post
	e.GET("/post/:post_id", ViewSinglePost)

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
	e.GET("/godmode.exe", func(c echo.Context) error {
		return nil
		i := identity.New()
		i.Powers = 95
		i.Sign()
		encoded, err := i.EncodeBase64()
		if err != nil {
			return c.String(http.StatusInternalServerError, "could not send your new id")
		}
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded})
		return c.String(http.StatusOK, encoded)
	})
	e.GET("/nuke.exe", func(c echo.Context) error {
		i := whoami(c)
		if i.Powers != 95 {
			return e404(c)
		}
		start := time.Now()
		forum.Nuke()
		return c.String(http.StatusOK, fmt.Sprintf("forum nuking took %v ", time.Since(start)))
	})

	e.GET("/reindex.exe", func(c echo.Context) error {
		i := whoami(c)
		if i.Powers != 95 {
			return e404(c)
		}
		start := time.Now()
		forum.RebuildIndex()
		return c.String(http.StatusOK, fmt.Sprintf("took %v ", time.Since(start)))
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

	// isso era pra mandar só JSON, mas agora manda HTML.
	// se eu quiser fazer um app, vou precisar voltar pra teoria do json.
	api := e.Group("visualbasic.exe")

	// view all topics
	api.GET("/json/:page", func(c echo.Context) error {
		page, err := strconv.ParseInt(c.Param("page"), 10, 32)
		if err != nil {
			return c.HTML(400, "wtf")
		}
		return c.JSON(http.StatusOK, forum.GetTopics(int(page), 10))
	})
	// view a single topic
	api.GET("/json/post/:topic_id", func(c echo.Context) error {
		topic_id := c.Param("topic_id")

		resultado, err := forum.ReadPost(topic_id)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		return c.JSON(http.StatusOK, resultado)
	})

	// view a single topic
	api.GET("/json/:topic_id", func(c echo.Context) error {
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
			return c.String(403, "you cannot create topic sir")
		}

		nova_conversa := forum.Post{}
		nova_conversa.CreatorId = identity.Id
		nova_conversa.IP = identity.IP

		if err := c.Bind(&nova_conversa); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		id, err := forum.CreateTopic(nova_conversa)
		nova_conversa.Id = id

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.HTML(200, RenderTemplate("partials/post", R{"Identity": identity, "Post": nova_conversa}))
	})

	api.GET("/post/:post_id/edit", FormEditPost)
	api.PUT("/post/:post_id", EditPost)

	// super estranho, mas ok.
	api.GET("/post/:post_id/flag", FormFlagPost)
	api.POST("/post/:post_id/flag", EditPost)

	// reply a thread
	api.POST("/post/:topic_id", ReplyThread)

	// freeze/unfreeze a post
	api.PUT("/post/:post_id/freeze", FreezePost)
	api.PUT("/post/:post_id/unfreeze", UnfreezePost)

	if err := e.Start(":3000"); err != nil {
		panic(err)
	}
}
