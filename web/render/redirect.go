package render

import (
	"fmt"
	"net/http"
)

type Redirect struct {
	Code     int
	Location string
	Request  *http.Request
}

func (s *Redirect) writeString(w http.ResponseWriter, location string) (err error) {
	s.WriteContentType(w)
	if (s.Code < http.StatusMultipleChoices || s.Code > http.StatusPermanentRedirect) && s.Code != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", s.Code))
	}
	http.Redirect(w, s.Request, location, s.Code)
	return nil
}

func (s *Redirect) Render(w http.ResponseWriter) error {
	return s.writeString(w, s.Location)
}

func (s *Redirect) WriteContentType(w http.ResponseWriter) {
}
