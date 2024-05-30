package api

import (
	"errors"
	"log"
	"net/http"
	"strconv"

	"github.com/Alfrederson/hilos/identity"

	"github.com/labstack/echo/v4"
)

func paramBool(c echo.Context, field string, def bool) bool {
	val, err := strconv.ParseBool(c.QueryParam(field))
	if err != nil {
		return def
	}
	return val
}
func paramInt(c echo.Context, field string, def int64) int64 {
	val, err := strconv.ParseInt(c.QueryParam(field), 32, 10)
	if err != nil {
		return def
	}
	return val
}

func requestedPage(c echo.Context) int64 {
	page, _ := strconv.ParseInt(c.QueryParam("p"), 32, 10)
	if page < 0 {
		page = 0
	}
	return page
}

// cria uma identidade nova e enfia no context.
func newIdentity(c echo.Context) *identity.Identity {
	i := identity.NewAnonymous()
	stuffIdentity(&i, c)
	return &i
}

// enfia uma identidade no context.
func stuffIdentity(i *identity.Identity, c echo.Context) (string, error) {
	i.IP = c.RealIP()
	i.Sign()
	encoded, err := i.EncodeBase64()
	if err != nil {
		return "", errors.New("vaca foi pro brejo")
	} else {
		c.SetCookie(&http.Cookie{Name: "rwt", Value: encoded, Path: "/"})
		return encoded, nil
	}
}

// verifica quem é o usuário logado.
// se não for ninguém, gera um usuário anônimo.
func whoami(c echo.Context) *identity.Identity {
	rwt, err := c.Cookie("rwt")
	if err != nil {
		log.Println("erro vendo o cookie:", err)
		return newIdentity(c)
	}
	i, err := identity.DecodeBase64(rwt.Value)
	if err != nil {
		log.Println("erro decodando:", err)
		return newIdentity(c)
	}
	if !i.Check() {
		log.Println("não passou na checagem")
		return newIdentity(c)
	}
	i.IP = c.RealIP()
	return i
}

func mustHave[T any](c echo.Context, key string) *T {
	thing := c.Get(key)
	if coisa, ok := thing.(*T); ok {
		return coisa
	} else {
		return nil
	}
}
