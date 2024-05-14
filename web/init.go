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
	*treeNode
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:           name,
		handlerFuncMap: make(map[string]map[string]handlerFunc),
		treeNode: &treeNode{
			val:        "/",
			children:   make([]*treeNode, 0),
			routerName: "/",
		},
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

	r.treeNode.Put(name)
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
	for _, group := range e.routerGroups {
		routerName := SubStringLast(r.RequestURI, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil {
			handler, ok := group.handlerFuncMap[node.routerName][method]
			if !ok {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "%s %s NOT ALLOWD", r.RequestURI, method)
				return
			}
			handler(&Context{
				W: w,
				R: r,
			})
			return
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
