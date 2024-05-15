package web

import (
	"fmt"
	"net/http"
)

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

type routerGroup struct {
	name           string
	handlerFuncMap map[string]map[string]HandlerFunc
	treeNode       *treeNode
	Middlewares    []MiddlewareFunc
}

func (g *routerGroup) MiddlewareHandle(middlewareFunc ...MiddlewareFunc) {
	g.Middlewares = append(g.Middlewares, middlewareFunc...)
}

func (g *routerGroup) MethodHandle(handlerFunc HandlerFunc, ctx *Context) {
	// 前置中间件
	if len(g.Middlewares) > 0 {
		for _, Middleware := range g.Middlewares {
			handlerFunc = Middleware(handlerFunc)
		}
	}

	handlerFunc(ctx)
}

func (r *routerGroup) handle(name string, method string, _handlerFunc HandlerFunc) {
	_, ok := r.handlerFuncMap[name]
	if !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
	}
	r.handlerFuncMap[name][method] = _handlerFunc

	r.treeNode.Put(name)
}

// Get get请求方式
func (r *routerGroup) Get(name string, _handlerFunc HandlerFunc) {
	r.handle(name, http.MethodGet, _handlerFunc)
}

func (r *routerGroup) Post(name string, _handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPost, _handlerFunc)
}

func (r *routerGroup) Put(name string, _handlerFunc HandlerFunc) {
	r.handle(name, http.MethodPut, _handlerFunc)
}

func (r *routerGroup) Delete(name string, _handlerFunc HandlerFunc) {
	r.handle(name, http.MethodDelete, _handlerFunc)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:           name,
		handlerFuncMap: make(map[string]map[string]HandlerFunc),
		treeNode: &treeNode{
			val:        "/",
			children:   make([]*treeNode, 0),
			routerName: "/",
		},
		Middlewares: make([]MiddlewareFunc, 0),
	}

	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
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

func (e *Engine) HTTPRequestHandler(w http.ResponseWriter, r *http.Request) {
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
			ctx := &Context{
				W: w,
				R: r,
			}
			group.MethodHandle(handler, ctx)

			return
		}

	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s NOT FOUND", r.URL.Path)
	return
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	e.HTTPRequestHandler(w, r)
}

func (e *Engine) Run() {

	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
