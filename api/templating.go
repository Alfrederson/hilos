package api

import (
	"bytes"
	"html/template"
)

func RenderTemplate(templateName string, data R) string {
	tmpl, err := template.ParseFiles("web/" + templateName + ".html")
	if err != nil {
		return "template dun gufd"
	}
	buffer := bytes.Buffer{}
	if err := tmpl.Execute(&buffer, data); err != nil {
		return err.Error()
	}
	return buffer.String()
}
