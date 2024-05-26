package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data any
}

const jsonContentType = "application/json;charset=uft-8"

func (s *Json) writeString(w http.ResponseWriter, data any, code int) (err error) {
	s.WriteContentType(w)
	w.WriteHeader(code)
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataJson)
	return err
}

func (s *Json) Render(w http.ResponseWriter, code int) error {
	return s.writeString(w, s.Data, code)
}

func (s *Json) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
