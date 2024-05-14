package web

import (
	"fmt"
	"net/http"
)

type routerGroup struct {
	name string

	/**
	key1: name
	key2: method
	value: func
	*/
	handlerFuncMap map[string]map[string]handlerFunc
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:           name,
		handlerFuncMap: make(map[string]map[string]handlerFunc),
	}

	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
}

type handlerFunc func(ctx *Context)

type router struct {
	routerGroups []*routerGroup
}

func (r *routerGroup) handle(name string, method string, _handlerFunc handlerFunc) {
	_, ok := r.handlerFuncMap[name]
	if !ok {
		r.handlerFuncMap[name] = make(map[string]handlerFunc)
	}
	r.handlerFuncMap[name][method] = _handlerFunc
}

// Get get请求方式
func (r *routerGroup) Get(name string, _handlerFunc handlerFunc) {
	r.handle(name, http.MethodGet, _handlerFunc)
}

func (r *routerGroup) Post(name string, _handlerFunc handlerFunc) {
	r.handle(name, http.MethodPost, _handlerFunc)
}

func (r *routerGroup) Put(name string, _handlerFunc handlerFunc) {
	r.handle(name, http.MethodPut, _handlerFunc)
}

func (r *routerGroup) Delete(name string, _handlerFunc handlerFunc) {
	r.handle(name, http.MethodDelete, _handlerFunc)
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router: router{
			routerGroups: make([]*routerGroup, 0),
		},
	}
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, routerGroup := range e.routerGroups {
		for name, m := range routerGroup.handlerFuncMap {
			uri := "/" + routerGroup.name + name
			if uri == r.RequestURI {
				// 找到了请求方式对应的url
				for _method, handler := range m {
					if _method == method {
						ctx := &Context{
							W: w,
							R: r,
						}
						handler(ctx)
						return
					}
				}
				// url匹配了但是方法不匹配
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "%s %s not allowd", uri, method)
				return
			}
		}
	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s NOT FOUND", r.URL.Path)
	return
}

func (e *Engine) Run() {
	//// 路由和功能的映射
	//for _, group := range e.routerGroups {
	//	for key, value := range group.handlerFuncMap {
	//		http.HandleFunc("/"+group.name+key, value)
	//	}
	//}

	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
