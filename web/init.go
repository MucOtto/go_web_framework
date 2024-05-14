package web

import "net/http"

type routerGroup struct {
	name           string
	handlerFuncMap map[string]handlerFunc
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:           name,
		handlerFuncMap: make(map[string]handlerFunc),
	}

	r.routerGroups = append(r.routerGroups, routerGroup)
	return routerGroup
}

type handlerFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	routerGroups []*routerGroup
}

func (r *routerGroup) Add(name string, handlerFunc handlerFunc) {
	r.handlerFuncMap[name] = handlerFunc
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

func (e *Engine) Run() {
	// 路由和功能的映射
	for _, group := range e.routerGroups {
		for key, value := range group.handlerFuncMap {
			http.HandleFunc("/"+group.name+key, value)
		}
	}

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
