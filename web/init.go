package web

import "net/http"

type handlerFunc func(w http.ResponseWriter, r *http.Request)

type router struct {
	handlerFuncMap map[string]handlerFunc
}

func (r *router) Add(name string, handlerFunc handlerFunc) {
	r.handlerFuncMap[name] = handlerFunc
}

type Engine struct {
	router
}

func New() *Engine {
	return &Engine{
		router: router{
			handlerFuncMap: make(map[string]handlerFunc),
		},
	}
}

func (e *Engine) Run() {
	// 路由和功能的映射
	for key, value := range e.handlerFuncMap {
		http.HandleFunc(key, value)
	}

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		return
	}
}
