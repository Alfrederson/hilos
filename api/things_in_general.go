package api

import (
	"hilos/identity"
	"log"
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
)

func requestedPage(c echo.Context) int64 {
	page, _ := strconv.ParseInt(c.QueryParam("p"), 32, 10)
	if page < 0 {
		page = 0
	}
	return page
}

func newIdentity(c echo.Context) *identity.Identity {
	i := identity.New()
	i.IP = c.RealIP()
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
	i.IP = c.Request().Header.Get("X-Forwarded-For")
	return i
}
