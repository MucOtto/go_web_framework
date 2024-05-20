package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/MucOtto/web/render"
	"html/template"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"reflect"
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

// Json  前后端Json格式获取解析
func (c *Context) Json(obj any) error {
	body := c.R.Body
	if body == nil {
		return errors.New("invalid request")
	}
	decoder := json.NewDecoder(body)

	// 是否允许传入的有多余未包含的字段
	if c.DisallowUnknownFields {
		decoder.DisallowUnknownFields()
	}

	// 是否允许少传入包含的字段
	c.DisallowLessFiles = true
	if c.DisallowLessFiles {
		err := validIsLessFields(obj, decoder)
		if err != nil {
			return err
		}
	}

	err := decoder.Decode(obj)
	if err != nil {
		return err
	}
	return nil
}

func validIsLessFields(obj any, decoder *json.Decoder) error {
	// 判断类型
	value := reflect.ValueOf(obj)
	if value.Kind() != reflect.Pointer {
		return errors.New("This argument need a pointer ")
	}

	// 这里取得了指针所指的对象
	elem := value.Elem().Interface()
	// 获取对象的类别
	valueOfElem := reflect.ValueOf(elem)

	switch valueOfElem.Kind() {
	case reflect.Struct:
		// 创建一个map存储key和elem里的字段进行比较
		m := make(map[string]any)
		err := decoder.Decode(&m)
		if err != nil {
			return err
		}

		for i := 0; i < valueOfElem.NumField(); i++ {
			field := valueOfElem.Type().Field(i)
			tag := field.Tag.Get("json")
			mapValue := m[tag]
			if mapValue == nil {
				return errors.New(fmt.Sprintf("filed [%s] is not exist", tag))
			}
		}
		// 因为decoder里面的数据流已经读完 不能二次重复读
		marshal, err := json.Marshal(m)
		if err != nil {
			log.Println(err)
		}
		err = json.Unmarshal(marshal, obj)
		if err != nil {
			return err
		}

	default:
		_ = decoder.Decode(obj)
	}
	return nil
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
