package web

import (
	"errors"
	"github.com/MucOtto/web/binding"
	"github.com/MucOtto/web/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const defaultMemory = 64 << 20

type Context struct {
	W                     http.ResponseWriter
	R                     *http.Request
	engine                *Engine
	queryCache            url.Values
	formCache             url.Values
	DisallowUnknownFields bool
	DisallowLessFiles     bool
}

// BindJson 前后端Json格式获取解析
func (c *Context) BindJson(obj any) error {
	jsonBinding := &binding.JSON
	jsonBinding.DisallowUnknownFields = c.DisallowUnknownFields
	jsonBinding.DisallowLessFiles = c.DisallowLessFiles
	return c.BindWith(obj, jsonBinding)
}

func (c *Context) BindXml(obj any) error {
	xmlBinding := &binding.XMl
	return c.BindWith(obj, xmlBinding)
}

func (c *Context) get(key string, cache url.Values) (map[string]string, bool) {
	dict, exist := make(map[string]string), false
	for k, v := range cache {
		if i := strings.IndexByte(k, '['); i >= 1 && k[:i] == key {
			if j := strings.IndexByte(k[i+1:], ']'); j >= 1 {
				exist = true
				dict[k[i+1:][:j]] = v[0]
			}
		}
	}
	return dict, exist
}

func (c *Context) GetMapQuery(key string) (map[string]string, bool) {
	c.initQueryCache()
	return c.get(key, c.queryCache)
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

func (c *Context) GetMapForm(key string) (map[string]string, bool) {
	c.initFormCache()
	return c.get(key, c.formCache)
}

func (c *Context) GetForm(key string) any {
	c.initFormCache()
	return c.formCache.Get(key)
}

func (c *Context) GetFormArray(key string) (values []string, ok bool) {
	c.initFormCache()
	values, ok = c.formCache[key]
	return values, ok
}

func (c *Context) initFormCache() {
	if c.R != nil {
		if err := c.R.ParseMultipartForm(defaultMemory); err != nil {
			// 是否发生了未上传文件之外的其他错误
			if !errors.Is(err, http.ErrNotMultipart) {
				log.Println(err)
			}
		}
		c.formCache = c.R.PostForm
	} else {
		c.formCache = url.Values{}
	}
}

// FormFile 处理表单上传文件
func (c *Context) FormFile(key string) (*multipart.FileHeader, error) {
	if err := c.R.ParseMultipartForm(defaultMemory); err != nil {
		return nil, err
	}
	file, header, err := c.R.FormFile(key)
	if err != nil {
		return nil, err
	}
	defer func(file multipart.File) {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
	}(file)
	return header, nil
}

func (c *Context) SaveAndUploadFile(file multipart.File, dst string) error {
	out, err := os.Create(dst)
	if err != nil {
		return err
	}

	defer func() {
		err := file.Close()
		if err != nil {
			log.Println(err)
		}
		err = out.Close()
		if err != nil {
			log.Println(err)
		}
	}()

	_, err = io.Copy(out, file)
	if err != nil {
		return err
	}

	return nil
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

func (c *Context) BindWith(obj any, bind binding.Binding) error {
	if err := bind.Bind(c.R, obj); err != nil {
		c.W.WriteHeader(http.StatusBadRequest)
		return err
	}
	return nil
}
