package web

import (
	"encoding/json"
	"fmt"
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

// Redirect 重定向
func (c *Context) Redirect(status int, location string) {
	if (status < http.StatusMultipleChoices || status > http.StatusPermanentRedirect) && status != http.StatusCreated {
		panic(fmt.Sprintf("Cannot redirect with status code %d", status))
	}
	http.Redirect(c.W, c.R, location, status)
}

func (c *Context) String(status int, format string, values ...any) (err error) {
	plainContentType := "text/plain; charset=utf-8"
	c.W.Header().Set("Content-Type", plainContentType)
	c.W.WriteHeader(status)
	if len(values) > 0 {
		_, err = fmt.Fprintf(c.W, format, values...)
		return
	}
	_, err = c.W.Write(StringToBytes(format))
	return
}
