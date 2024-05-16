package web

import (
	"encoding/json"
	"net/http"
	"net/url"
)

type Context struct {
	W      http.ResponseWriter
	R      *http.Request
	engine *Engine
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

func (c *Context) HTMLTemplate(name string, data any) error {
	c.W.Header().Set("Content-Type", "text/html;charset=uft-8")
	err := c.engine.HTMLRender.Template.ExecuteTemplate(c.W, name, data)
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) JsonTemplate(data any) error {
	c.W.Header().Set("Content-Type", "application/json;charset=uft-8")
	c.W.WriteHeader(http.StatusOK)
	dataJson, err := json.Marshal(data)
	if err != nil {
		return err
	}
	_, err = c.W.Write(dataJson)
	return err
}

func (c *Context) FileAttachment(filepath, filename string) {
	if isASCII(filename) {
		c.W.Header().Set("Content-Disposition", `attachment; filename="`+filename+`"`)
	} else {
		c.W.Header().Set("Content-Disposition", `attachment; filename*=UTF-8''`+url.QueryEscape(filename))
	}
	http.ServeFile(c.W, c.R, filepath)
}

func (c *Context) FileFromFS(filepath string, fs http.FileSystem) {
	defer func(old string) {
		c.R.URL.Path = old
	}(c.R.URL.Path)

	c.R.URL.Path = filepath

	http.FileServer(fs).ServeHTTP(c.W, c.R)
}
