package web

import (
	"fmt"
	"github.com/MucOtto/web/render"
	"html/template"
	"net/http"
	"sync"
)

type HandlerFunc func(ctx *Context)

type MiddlewareFunc func(handlerFunc HandlerFunc) HandlerFunc

type routerGroup struct {
	name              string
	handlerFuncMap    map[string]map[string]HandlerFunc
	middlewareFuncMap map[string]map[string][]MiddlewareFunc // 中间件和路由的映射
	treeNode          *treeNode
	Middlewares       []MiddlewareFunc
}

func (g *routerGroup) MiddlewareHandle(middlewareFunc ...MiddlewareFunc) {
	g.Middlewares = append(g.Middlewares, middlewareFunc...)
}

func (g *routerGroup) MethodHandle(name string, method string, handlerFunc HandlerFunc, ctx *Context) {
	// 所有方法适用的中间件
	if len(g.Middlewares) > 0 {
		for _, Middleware := range g.Middlewares {
			handlerFunc = Middleware(handlerFunc)
		}
	}

	// 各自路由对应的中间件
	middlewareFuncs, ok := g.middlewareFuncMap[name][method]
	if ok && len(middlewareFuncs) > 0 {
		for _, Middleware := range middlewareFuncs {
			handlerFunc = Middleware(handlerFunc)
		}
	}

	handlerFunc(ctx)
}

func (r *routerGroup) handle(name string, method string, _handlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	_, ok := r.handlerFuncMap[name]
	if !ok {
		r.handlerFuncMap[name] = make(map[string]HandlerFunc)
		r.middlewareFuncMap[name] = make(map[string][]MiddlewareFunc)
	}
	r.handlerFuncMap[name][method] = _handlerFunc
	r.middlewareFuncMap[name][method] = append(r.middlewareFuncMap[name][method], middlewares...)

	r.treeNode.Put(name)
}

// Get get请求方式
func (r *routerGroup) Get(name string, _handlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	r.handle(name, http.MethodGet, _handlerFunc, middlewares...)
}

func (r *routerGroup) Post(name string, _handlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	r.handle(name, http.MethodPost, _handlerFunc, middlewares...)
}

func (r *routerGroup) Put(name string, _handlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	r.handle(name, http.MethodPut, _handlerFunc, middlewares...)
}

func (r *routerGroup) Delete(name string, _handlerFunc HandlerFunc, middlewares ...MiddlewareFunc) {
	r.handle(name, http.MethodDelete, _handlerFunc, middlewares...)
}

type router struct {
	routerGroups []*routerGroup
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:              name,
		handlerFuncMap:    make(map[string]map[string]HandlerFunc),
		middlewareFuncMap: make(map[string]map[string][]MiddlewareFunc),
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
	funcMap    template.FuncMap
	HTMLRender *render.HTMLRender
	pool       sync.Pool
}

func New() *Engine {
	engine := &Engine{
		router: router{
			routerGroups: make([]*routerGroup, 0),
		},
	}
	engine.pool.New = func() any {
		return engine.allocateContext()
	}
	return engine
}

func (e *Engine) allocateContext() any {
	return &Context{engine: e}
}

func (e *Engine) SetFuncMap(funcMap template.FuncMap) {
	e.funcMap = funcMap
}

func (e *Engine) LoadTemplate(pattern string) {
	t := template.Must(template.New("").Funcs(e.funcMap).ParseGlob(pattern))
	e.HTMLRender = &render.HTMLRender{
		Template: t,
	}
}

func (e *Engine) HTTPRequestHandler(context *Context, w http.ResponseWriter, r *http.Request) {
	method := r.Method
	for _, group := range e.routerGroups {
		routerName := SubStringLast(r.URL.Path, "/"+group.name)
		node := group.treeNode.Get(routerName)
		if node != nil {
			handler, ok := group.handlerFuncMap[node.routerName][method]
			if !ok {
				w.WriteHeader(http.StatusMethodNotAllowed)
				fmt.Fprintf(w, "%s %s NOT ALLOWD", r.RequestURI, method)
				return
			}
			group.MethodHandle(node.routerName, method, handler, context)

			return
		}

	}
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s NOT FOUND", r.URL.Path)
	return
}

func (e *Engine) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	context := e.pool.Get().(*Context)
	context.W = w
	context.R = r
	e.HTTPRequestHandler(context, w, r)
	e.pool.Put(context)
}

func (e *Engine) Run() {

	http.Handle("/", e)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
