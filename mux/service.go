package mux

import (
	"net/http"
	"reflect"
	"strings"

	"go.shu.run/bootstrap/utils/namec"
)

//Service http server service
type Service struct {
	name         string
	initHandlers []func(*Mux)
	middleware   []MiddlewareFunc
	handlers     []string
	handlerMap   map[string]HandlerFunc
}

//Apply Apply
func (srv *Service) Apply(mux *Mux) {
	router := mux
	if srv.name != "" {
		router = router.Group("/" + srv.name)
	}

	router.Use(srv.middleware...)

	for _, handle := range srv.initHandlers {
		handle(router)
	}

	for _, name := range srv.handlers {
		method, path := findMethod(name)
		if method != "" {
			router.Handle(method, "/"+path, srv.handlerMap[name])
		} else {
			router.GetPost("/"+path, srv.handlerMap[name])
		}
	}
}

func newService(service interface{}) *Service {
	var (
		refV = reflect.ValueOf(service)
		refT = refV.Type()
		srv  = &Service{
			name:       "",
			handlerMap: make(map[string]HandlerFunc),
		}
	)

	for i := 0; i < refT.NumMethod(); i++ {
		name := refT.Method(i).Name
		if 'A' <= name[0] && name[0] <= 'Z' {
			switch method := refV.Method(i).Interface().(type) {
			case func(*C) R:
				name = namec.SnakeCase(name)
				if _, find := srv.handlerMap[name]; !find {
					srv.handlers = append(srv.handlers, name)
				}
				srv.handlerMap[name] = method
			case func(next HandlerFunc) HandlerFunc:
				srv.middleware = append(srv.middleware, method)
			case func(*Mux):
				srv.initHandlers = append(srv.initHandlers, method)
			case func() string:
				if name == "Name" {
					srv.name = method()
				}
			}
		}
	}

	return srv
}

//HandleService HandleService
func (mux *Mux) HandleService(service interface{}) {
	mux.Infof("注册服务: %s", reflect.Indirect(reflect.ValueOf(service)).Type().Name())
	newService(service).Apply(mux)
}

func findMethod(name string) (method, path string) {
	for _, method := range serviceMethod {
		if strings.HasPrefix(name, strings.ToLower(method)+"_") {
			return method, name[len(method)+1:]
		}
	}
	return "", name
}

var serviceMethod = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
	http.MethodConnect,
	http.MethodOptions,
	http.MethodTrace,
}
