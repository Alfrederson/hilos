package api

import (
	"bytes"
	"errors"
	"fmt"
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

// copiado e colado..
func m(pairs ...any) (map[string]any, error) {
	if len(pairs)%2 != 0 {
		return nil, errors.New("misaligned map")
	}

	result := make(map[string]any, len(pairs)/2)

	for i := 0; i < len(pairs); i += 2 {
		key, ok := pairs[i].(string)
		if !ok {
			return nil, fmt.Errorf("cannot use type %T as map key", pairs[i])
		}
		result[key] = pairs[i+1]
	}
	return result, nil
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

// se tem o HTML e o nome é Pugger, claramente não deu certo usar o Pug.
type Pugger struct {
	engine *html.Engine
}

func MakePugger() Pugger {
	p := Pugger{}
	p.engine = html.New("./web3", ".html")
	p.engine.Reload(true)
	p.engine.AddFunc("crop", crop)
	p.engine.AddFunc("m", m)
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
				"crop":       crop,
				"formatTime": formatTime,
				"m":          m,
			}).Parse(readFile("web3/" + templateName + ".html"))

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
