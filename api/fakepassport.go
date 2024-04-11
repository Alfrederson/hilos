package api

import (
	"errors"
	"hilos/identity"
	"hilos/util"
	"log"
	"net/http"

	"github.com/labstack/echo/v4"
)

// WTF: a gente deveria enfiar a referência para o session inteiro paraa cada template.
func Fakepassport_Options(c echo.Context) error {
	s := session(c)
	return RenderPartial(c, "passport/options", R{"Identity": s.id})
}

// formulário para criar o passaporte falso
func Get_Fakepassport_New(c echo.Context) error {
	s := session(c)
	return RenderPartial(c, "passport/new", R{"Identity": s.id})
}

// bota um cookie com o passaporte falso e redireciona, ou
// retorna form para o passaporte falso caso dê errado.
func Post_Fakepassport_New(c echo.Context) error {

	type FakepassportBody struct {
		Name string `form:"name"`
	}

	b := FakepassportBody{}
	c.Bind(&b)
	errorMessage := ""
	encoded := ""
	var i identity.Identity

	if !util.IsAlphaNumeric(b.Name) {
		errorMessage = "only a-z, 0-9 allowed sir"
		goto fail
	}
	if len(b.Name) < 3 {
		errorMessage = "name must be at least 3 letters sir"
		goto fail
	}
	if len(b.Name) > 12 {
		errorMessage = "name must be at most 12 letters sir"
		goto fail
	}

	i = identity.NewNamed(b.Name)
	encoded, _ = stuffIdentity(&i, c)

	return RenderPartial(c, "passport/success", R{
		"Identity": i,
		"Passport": encoded,
	})

fail:
	return RenderPartial(c, "passport/new", R{
		"Error": errorMessage,
	})
}

func Post_Fakepassport_Use(c echo.Context) error {
	type FakepassportBody struct {
		Passport string `form:"passport"`
	}
	b := FakepassportBody{}
	var err error
	var encoded string
	var i *identity.Identity
	if err = c.Bind(&b); err != nil {
		goto fail
	}
	if i, err = identity.DecodeBase64(b.Passport); err != nil {
		log.Println("error decoding: ", err)
		goto fail
	}
	if !i.Check() {
		err = errors.New("signature invalid or fake")
		goto fail
	}
	if encoded, err = i.Sign().EncodeBase64(); err != nil {
		goto fail
	}
	c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded, Path: "/"})
	return RenderPartial(c, "passport/success", R{
		"Identity": i,
		"Passport": encoded,
	})
fail:
	return RenderPartial(c, "passport/use", R{
		"Error": err.Error(),
	})
}

func fakepassport(g *echo.Group) {
	g.GET("/options", Fakepassport_Options)

	g.GET("/new", Get_Fakepassport_New)
	g.POST("/new", Post_Fakepassport_New)

	g.GET("/use", func(c echo.Context) error {
		return RenderPartial(c, "passport/use", R{})
	})
	g.POST("/use", Post_Fakepassport_Use)

	g.GET("", func(c echo.Context) error {
		s := session(c)
		return c.Render(200, "fakepassport", R{
			"Identity": s.id,
		})
	})
}
