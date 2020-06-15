package mux

import (
    "context"
    "fmt"
    "net/http"
    "path/filepath"
    "strings"
    "sync"

    "go.shu.run/bootstrap/logger"
)

func New() *Mux {
    mux := &Mux{log: logger.Nil}
    mux.cPool.New = func() interface{} { return &C{} }
    return mux
}

type HandlerFunc func(c *C) R

type MiddlewareFunc func(next HandlerFunc) HandlerFunc

type Mux struct {
    root       *Mux
    prefix     string
    middleware []MiddlewareFunc
    log        logger.Logger

    trees       map[string]*node
    NotFound    http.Handler
    NotAllowed  http.Handler
    AllowOrigin []string

    server *http.Server

    cancel  context.CancelFunc
    closing bool
    cPool   sync.Pool
    locker  sync.Mutex
    wait    sync.WaitGroup
}

func (mux *Mux) Log() logger.Logger {
    return mux.getRoot().log
}

func (mux *Mux) SetLogger(l logger.Logger) {
    mux.getRoot().log = l
}

func (mux *Mux) Group(prefix string) *Mux {
    return &Mux{
        root:       mux.getRoot(),
        middleware: mux.middleware,
        prefix:     filepath.Clean(mux.prefix + prefix),
    }
}

func (mux *Mux) Use(middleware ...MiddlewareFunc) {
    mux.middleware = append(mux.middleware, middleware...)
}

func (mux *Mux) GET(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
    mux.Handle(http.MethodGet, path, handle, middleware...)
}

func (mux *Mux) POST(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
    mux.Handle(http.MethodPost, path, handle, middleware...)
}

func (mux *Mux) GetPost(path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
    mux.GET(path, handle, middleware...)
    mux.POST(path, handle, middleware...)
}

func (mux *Mux) Handle(method, path string, handle HandlerFunc, middleware ...MiddlewareFunc) {
    root := mux.getRoot()
    path = CleanPath(mux.prefix + path)
    mux.Log().Infof("注册路由: %-4s -> %s", method, path)

    if root.trees == nil {
        root.trees = make(map[string]*node)
    }

    mMap := root.trees[method]
    if mMap == nil {
        mMap = new(node)
        root.trees[method] = mMap
    }

    mMap.addRoute(path, mux.applyMw(handle, append(middleware, mux.middleware...)...))
}

func (mux *Mux) Reset() {
    router := mux.getRoot()
    if len(router.middleware) > 0 {
        router.middleware = router.middleware[:0]
    }

    if len(router.trees) > 0 {
        for key := range router.trees {
            delete(router.trees, key)
        }
    }
}

func (mux *Mux) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    if mux.closing {
        http.Error(w, "服务正在关闭...", 200)
        return
    }

    defer mux.handlePanic(w, req)

    path := req.URL.Path
    req.URL.Path = CleanPath(path)

    root := mux.getRoot()

    mux.handleOrigin(w, req)
    if req.Method == http.MethodOptions {
        w.WriteHeader(http.StatusNoContent)
        return
    }

    if !strings.HasSuffix(req.URL.Path, "/status") && !strings.HasSuffix(req.URL.Path, "/favicon.ico") {
        mux.Log().Infof("%-4s -> %s", req.Method, req.URL.Path)
    }

    if mTree := root.trees[req.Method]; mTree != nil {
        handle, ps, _ := mTree.getValue(path)
        if handle != nil {
            ctx := mux.newContext(req, ps)
            defer ctx.release()
            if resp := handle(ctx); resp != nil {
                resp.WriteTo(ctx, w)
            }
            return
        }
    }

    // HandlerFunc 404
    if root.NotFound != nil {
        //req.URL.Path = req.URL.Path[1:]
        root.NotFound.ServeHTTP(w, req)
    } else {
        http.NotFound(w, req)
    }
}

func (mux *Mux) getRoot() *Mux {
    if mux.root == nil {
        return mux
    }
    return mux.root
}

func (mux *Mux) applyMw(handlerFunc HandlerFunc, middleware ...MiddlewareFunc) HandlerFunc {
    for i := len(middleware) - 1; i >= 0; i-- {
        handlerFunc = middleware[i](handlerFunc)
    }
    return handlerFunc
}

func (mux *Mux) handlePanic(w http.ResponseWriter, _ *http.Request) {
    if err := recover(); err != nil {
        mux.Log().Errorf("http error: %v", err)
        http.Error(w, fmt.Sprintf("%v", err), 500)
        panic(err)
    }
}

func (mux *Mux) newContext(req *http.Request, pathParams map[string]string) *C {
    mux.locker.Lock()
    defer mux.locker.Unlock()

    c := mux.cPool.Get().(*C)

    c.Log = mux.Log()
    c.mux = mux.getRoot()
    c.Request = req

    if len(c.Values) > 0 {
        for key := range c.Values {
            delete(c.Values, key)
        }
    }

    if len(pathParams) > 0 {
        if c.Values == nil {
            c.Values = make(map[string]interface{}, len(pathParams))
        }
        for key, val := range pathParams {
            c.Values[pathParamKey+key] = val
        }
    }

    return c
}

func (mux *Mux) handleOrigin(w http.ResponseWriter, req *http.Request) {
    if len(mux.AllowOrigin) == 0 {
        return
    }

    origin := req.Header.Get("Origin")
    if origin == "" {
        return
    }

    var match bool
    for _, allow := range mux.AllowOrigin {
        if match, _ = filepath.Match(allow, origin); match {
            mux.Log().Debugf("match origin: '%s' by '%s'", origin, allow)
            break
        }
    }

    if method := req.Header.Get("Access-Control-Request-Method"); method != "" {
        w.Header().Set("Access-Control-Allow-Methods", method)
    }

    if headers := req.Header.Get("Access-Control-Request-Headers"); headers != "" {
        w.Header().Set("Access-Control-Allow-Headers", headers)
    }

    w.Header().Set("Access-Control-Allow-Origin", origin)
    w.Header().Set("Access-Control-Allow-Credentials", "true")
    w.Header().Set("Access-Control-Max-Age", "86400")
}
