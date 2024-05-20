package web

import (
	"github.com/MucOtto/web/render"
	"html/template"
	"net/http"
	"net/url"
	"strings"
)

type Context struct {
	W          http.ResponseWriter
	R          *http.Request
	engine     *Engine
	queryCache url.Values
}

func (c *Context) GetMapQuery(key string) (map[string]string, bool) {
	c.initQueryCache()
	dict, exist := make(map[string]string), false
	for k, v := range c.queryCache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dict[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dict, exist
}

func (c *Context) GetQuery(key string) any {
	c.initQueryCache()
	return c.queryCache.Get(key)
}

func (c *Context) GetQueryArray(key string) (values []string, ok bool) {
	c.initQueryCache()
	values, ok = c.queryCache[key]
	return
}

func (c *Context) initQueryCache() {
	if c.R != nil {
		c.queryCache = c.R.URL.Query()
	} else {
		c.queryCache = make(url.Values)
	}
}

func (c *Context) HTMLTemplate(name string, data any) error {
	err := c.Render(http.StatusOK, &render.HTML{
		Template: template.New("temp"),
		Name:     name,
		Data:     data,
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *Context) JsonTemplate(data any) error {
	err := c.Render(http.StatusOK, &render.Json{
		Data: data,
	})
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
	c.Render(http.StatusOK, &render.Redirect{
		Code:     status,
		Location: location,
		Request:  c.R,
	})
}

func (c *Context) String(status int, format string, values ...any) (err error) {
	err = c.Render(status, &render.String{
		Format: format,
		Data:   values,
	})
	return err
}

func (c *Context) Render(code int, r render.Render) error {
	err := r.Render(c.W)
	if code != http.StatusOK {
		c.W.WriteHeader(code)
	}
	return err
}
