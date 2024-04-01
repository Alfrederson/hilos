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

func onlyCopsAllowed(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		id := whoami(c)
		if id.Powers != 95 {
			return c.String(http.StatusUnavailableForLegalReasons, "not found")
		}
		c.Set("id", id)
		return next(c)
	}
}

func identifier(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		s := session(c)
		// coloca o id na sessão a partir do cookie...
		s.id = whoami(c)
		// aqui a gente vai carregar algumas coisas do arquivo,
		// tipo mensagens privadas e tal. eu acho.

		return next(c)
	}
}

func Start() {
	e := echo.New()

	pugger := MakePugger()
	e.Renderer = pugger

	//	e.Renderer = t
	e404 := func(c echo.Context) error {
		return c.String(404, "not found")
	}
	ç := func(c echo.Context) error {
		log.Println(c.RealIP(), " looking for ", c.Request().RequestURI)
		return c.String(200, "🤡")
	}

	// coisas para enganar alguns bots
	e.GET("/favicon.ico", e404)
	e.GET("/robots.txt", func(c echo.Context) error {
		log.Println("🕷️ ", c.RealIP())
		return c.String(200, "User-agent: *\nDisallow:")
	})
	e.GET("/admin.php", ç)
	e.GET("/wp-admin", ç)
	e.GET("/.env", ç)

	root := e.Group("", identifier)

	root.GET("/testCop", func(c echo.Context) error {
		return c.String(200, "You're a cop!")
	}, onlyCopsAllowed)

	root.GET("/", Index)
	root.GET("/chat", Chat)
	// View a thread/topic/post whatever
	root.GET("/:topic_id", ViewTopic)
	// View the last post
	root.GET("/new", ViewLastPost)

	// view all posts by a user
	root.GET("/by/:user_id", ViewUserPosts)
	// view a single post
	root.GET("/post/:post_id", ViewSinglePost)

	cop := root.Group("/cop", onlyCopsAllowed)
	// view all reports
	cop.GET("/reports", Cop_ViewReports)
	// dismiss a report
	cop.POST("/reports/:report_id/dismiss", Cop_DismissReport)
	// view all bans
	cop.GET("/bans", Cop_ViewBans)
	// nuke forum in case of legal stuff
	cop.GET("/nuke.exe", func(c echo.Context) error {
		start := time.Now()
		forum.Nuke()
		return c.String(http.StatusOK, fmt.Sprintf("forum nuking took %v ", time.Since(start)))
	})
	cop.GET("/reindex.exe", func(c echo.Context) error {
		i := whoami(c)
		if i.Powers != 95 {
			return e404(c)
		}
		start := time.Now()
		forum.RebuildIndex()
		return c.String(http.StatusOK, fmt.Sprintf("took %v ", time.Since(start)))
	})
	cop.GET("/godmode.exe", func(c echo.Context) error {
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

	// freeze/unfreeze a post
	cop.PUT("/post/:post_id/freeze", Cop_FreezePost)
	cop.PUT("/post/:post_id/unfreeze", Cop_UnfreezePost)

	// creates a new identity
	root.GET("/newidentity.exe", func(c echo.Context) error {
		i := identity.New()
		encoded, err := i.EncodeBase64()
		if err != nil {
			return c.String(http.StatusInternalServerError, "could not send your new id")
		}
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded})
		return c.String(http.StatusOK, encoded)
	})

	root.GET("fakepassport.exe/:passport", func(c echo.Context) error {
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

	root.GET("/whoami.exe", func(c echo.Context) error {
		user := whoami(c)
		return c.String(http.StatusOK, fmt.Sprint(user.Id, user.Name))
	})
	// isso era pra mandar só JSON, mas agora manda HTML.
	// se eu quiser fazer um app, vou precisar voltar pra teoria do json.
	api := e.Group("visualbasic.exe", identifier)

	// view all topics
	api.GET("/json/:page", func(c echo.Context) error {
		page, err := strconv.ParseInt(c.Param("page"), 10, 32)
		if err != nil {
			return c.HTML(400, "wtf")
		}
		return c.JSON(http.StatusOK, forum.GetRootTopics(int(page), 10))
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

	// create a root topic
	api.POST("/", func(c echo.Context) error {
		s := session(c)

		nova_conversa := forum.Post{}
		nova_conversa.CreatorId = s.id.Id
		nova_conversa.IP = s.id.IP

		if err := c.Bind(&nova_conversa); err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}

		id, err := forum.CreateRootTopic(nova_conversa)
		nova_conversa.Id = id

		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		return c.HTML(200, RenderTemplate("partials/post", R{"Identity": s.id, "Post": nova_conversa}))
	}, onlyCopsAllowed)

	api.GET("/post/:post_id/edit", FormEditPost)
	api.PUT("/post/:post_id", EditPost)

	// reportar um post
	api.GET("/post/:post_id/flag", FormFlagPost)
	api.POST("/post/:post_id/flag", FlagPost)

	// reply a thread
	api.POST("/post/:topic_id", ReplyThread)

	// hardcoded porque é a única porta do raspberry pi que não está quebrada
	if err := e.Start(":8999"); err != nil {
		panic(err)
	}
}
