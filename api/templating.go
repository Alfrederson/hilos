package api

import (
	"bytes"
	"html/template"
	"io/ioutil"
	"log"
)

func crop(text string) string {
	if len(text) > 29 {
		return text[:26] + "..."
	} else {
		return text
	}
}

func readFile(filename string) string {
	text, err := ioutil.ReadFile(filename)
	if err != nil {
		return err.Error()
	}
	return string(text)
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
