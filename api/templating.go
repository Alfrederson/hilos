package api

import (
	"bytes"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/gofiber/template/html/v2"
	"github.com/labstack/echo/v4"
)

func crop(text string) string {
	if len(text) > 29 {
		return text[:26] + "..."
	} else {
		return text
	}
}

func formatTime(t time.Time) string {
	return t.Format("15:04:03 01-Jan-2006")
}

func readFile(filename string) string {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return err.Error()
	}
	return string(text)
}

type Pugger struct {
	engine *html.Engine
}

func MakePugger() Pugger {
	p := Pugger{}
	p.engine = html.New("./web3", ".html")
	p.engine.Reload(true)
	p.engine.AddFunc("crop", crop)
	p.engine.AddFunc("formatTime", formatTime)
	return p
}

func (p Pugger) Render(w io.Writer, name string, data interface{}, c echo.Context) error {
	err := p.engine.Render(w, name, data, "layout/main")
	if err != nil {
		log.Println(err)
	}
	return err
}

// O certo é não fazer isso, hehe.
func RenderTemplate(templateName string, data R) string {
	tmpl, err :=
		template.
			New("t").
			Funcs(template.FuncMap{
				"crop": crop,
			}).Parse(readFile("web/" + templateName + ".html"))

	if err != nil {
		log.Println(err)
		return "template dun gufd"
	}

	buffer := bytes.Buffer{}
	if err := tmpl.Execute(&buffer, data); err != nil {
		return err.Error()
	}
	return buffer.String()
}
