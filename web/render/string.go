package render

import (
	"fmt"
	"github.com/MucOtto/web/internel/bytesconv"
	"net/http"
)

type String struct {
	Format string
	Data   []any
}

const contentType = "text/plain; charset=utf-8"

func (s *String) writeString(w http.ResponseWriter, format string, data []any) (err error) {
	s.WriteContentType(w)
	if len(data) > 0 {
		_, err = fmt.Fprintf(w, format, data...)
		return
	}
	_, err = w.Write(bytesconv.StringToBytes(format))
	return
}

func (s *String) Render(w http.ResponseWriter) error {
	return s.writeString(w, s.Format, s.Data)
}

func (s *String) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, contentType)
}
