package web

import (
	"fmt"
	"net/http"
)

type routerGroup struct {
	name           string
	handlerFuncMap map[string]handlerFunc

	/**
	key: http method
	value: url
	*/
	handlerMethodMap map[string][]string
}

func (r *router) Group(name string) *routerGroup {
	routerGroup := &routerGroup{
		name:             name,
		handlerFuncMap:   make(map[string]handlerFunc),
		handlerMethodMap: make(map[string][]string),
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

// Any 任何请求方式
func (r *routerGroup) Any(name string, handlerFunc handlerFunc) {
	r.handlerFuncMap[name] = handlerFunc
	r.handlerMethodMap["ANY"] = append(r.handlerMethodMap["ANY"], name)
}

// Get get请求方式
func (r *routerGroup) Get(name string, handlerFunc handlerFunc) {
	r.handlerFuncMap[name] = handlerFunc
	r.handlerMethodMap[http.MethodGet] = append(r.handlerMethodMap[http.MethodGet], name)
}

func (r *routerGroup) Post(name string, handlerFunc handlerFunc) {
	r.handlerFuncMap[name] = handlerFunc
	r.handlerMethodMap[http.MethodPost] = append(r.handlerMethodMap[http.MethodPost], name)
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
		for name, methodHandler := range group.handlerFuncMap {
			url := "/" + group.name + name
			if url == r.RequestURI {
				routers, ok := group.handlerMethodMap["ANY"]
				if ok {
					for _, routerName := range routers {
						if routerName == name {
							methodHandler(w, r)
							return
						}
					}
				}

				// 没有匹配到的话
				routers, ok = group.handlerMethodMap[method]
				if !ok {
					w.WriteHeader(http.StatusMethodNotAllowed)
					fmt.Fprintf(w, "%s %s not allowd \n", url, method)
				}
				for _, routerName := range routers {
					if routerName == name {
						methodHandler(w, r)
						return
					}
				}
			}
		}
	}
	// 都没匹配上 没有相应的方法
	w.WriteHeader(http.StatusNotFound)
	fmt.Fprintf(w, "%s not found\n", r.RequestURI)
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
