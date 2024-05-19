package render

import (
	"encoding/json"
	"net/http"
)

type Json struct {
	Data any
}

const jsonContentType = "application/json;charset=uft-8"

func (s *Json) writeString(w http.ResponseWriter, data any) (err error) {
	s.WriteContentType(w)
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = w.Write(dataJson)
	return err
}

func (s *Json) Render(w http.ResponseWriter) error {
	return s.writeString(w, s.Data)
}

func (s *Json) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, jsonContentType)
}
