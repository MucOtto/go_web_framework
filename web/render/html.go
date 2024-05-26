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

func (s *HTML) writeString(w http.ResponseWriter, name string, data any, code int) (err error) {
	s.WriteContentType(w)
	w.WriteHeader(code)
	err = s.Template.ExecuteTemplate(w, name, data)
	if err != nil {
		return err
	}
	return nil
}

func (s *HTML) Render(w http.ResponseWriter, code int) error {
	return s.writeString(w, s.Name, s.Data, code)
}

func (s *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, htmlContentType)
}
