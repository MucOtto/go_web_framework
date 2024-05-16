package web

import (
	"html/template"
	"net/http"
)

type Context struct {
	W http.ResponseWriter
	R *http.Request
}

func (c *Context) HTML(html string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=uft-8")
	c.W.WriteHeader(http.StatusOK)
	_, err := c.W.Write([]byte(html))
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) HTMLTemplate(name string, data any, pattern string) error {
	c.W.Header().Set("Content-Type", "text/html;charset=uft-8")
	t := template.New(name)
	t, err := t.ParseGlob(pattern)
	if err != nil {
		return err
	}
	err = t.Execute(c.W, data)
	if err != nil {
		return err
	}
	return nil
}
