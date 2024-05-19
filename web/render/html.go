package render

import (
	"html/template"
	"net/http"
)

type HTML struct {
	Template *template.Template
	Name     string
	Data     any
}
type HTMLRender struct {
	Template *template.Template
}

const htmlContentType = "text/html;charset=uft-8"

func (s *HTML) writeString(w http.ResponseWriter, name string, data any) (err error) {
	s.WriteContentType(w)
	err = s.Template.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *HTML) Render(w http.ResponseWriter) error {
	return s.writeString(w, s.Name, s.Data)
}

func (s *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
