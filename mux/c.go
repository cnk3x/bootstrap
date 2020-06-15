package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"sync"
	"time"

	"go.shu.run/bootstrap/logger"
	"go.shu.run/bootstrap/mux/binding"
	"go.shu.run/bootstrap/mux/formbind"
)

const (
	modelKey     = "__model__"
	pathParamKey = "__path_param__"

	mimeTypeJSON = "application/json; charset=utf-8"
)

var _ context.Context = (*C)(nil)

type C struct {
	*http.Request
	Values map[string]interface{} //中间件值载体

	Log logger.Logger
	mux *Mux
	sync.Mutex
}

func (c *C) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

func (c *C) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

func (c *C) Err() error {
	return c.Request.Context().Err()
}

func (c *C) Value(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}

func (c *C) release() {
	c.Request = nil
	if len(c.Values) > 0 {
		for key := range c.Values {
			delete(c.Values, key)
		}
	}
	c.mux.cPool.Put(c)
}

func (c *C) Set(name string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if c.Values == nil {
		c.Values = make(map[string]interface{}, 1)
	}
	c.Values[name] = value
}

func (c *C) Get(name string) interface{} {
	if len(c.Values) > 0 {
		return c.Values[name]
	}
	return nil
}

func (c *C) Param(name string) string {
	v, ok := c.Get(pathParamKey + name).(string)
	if !ok {
		return "-"
	}
	return v
}

func (c *C) ParamInt(name string) int64 {
	v, _ := strconv.ParseInt(c.Param(name), 10, 64)
	return v
}

func (c *C) HeaderGet(names ...string) string {
	for _, name := range names {
		s := c.Request.Header.Get(name)
		if s != "" {
			return s
		}
	}
	return ""
}

func (c *C) FormValue(names ...string) string {
	for _, name := range names {
		s := c.Request.FormValue(name)
		if s != "" {
			return s
		}
	}
	return ""
}

func (c *C) FormInt(names ...string) int64 {
	v, _ := strconv.ParseInt(c.FormValue(names...), 10, 64)
	return v
}

func (c *C) FormFloat(names ...string) float64 {
	v, _ := strconv.ParseFloat(c.FormValue(names...), 64)
	return v
}

func (c *C) FormBool(names ...string) bool {
	v, _ := strconv.ParseBool(c.FormValue(names...))
	return v
}

func (c *C) SetModel(value interface{}) {
	c.Set(modelKey, value)
}

func (c *C) Model() interface{} {
	return c.Get(modelKey)
}

func (c *C) GetHostURL() string {
	proto := c.Request.Header.Get("X-Forwarded-Proto")
	if proto == "" {
		if c.TLS != nil {
			proto = "https"
		} else {
			proto = "http"
		}
	}

	return fmt.Sprintf("%s://%s", proto, c.Host)
}

func (c *C) GetPageURI() string {
	if c.URL.IsAbs() {
		return c.URL.String()
	}
	return c.GetHostURL() + c.URL.String()
}

func (c *C) GetURI(format string, args ...interface{}) string {
	return c.GetHostURL() + fmt.Sprintf(format, args...)
}

func (c *C) GetForm() url.Values {
	switch c.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		_ = c.ParseForm()
		return c.Form
	case http.MethodGet:
		return c.URL.Query()
	default:
		return nil
	}
}

func (c *C) Bind(i interface{}) error {
	binder := binding.Default(c.Method, c.HeaderGet("Content-Type"))
	return binder.Bind(c.Request, i)
}

func (c *C) BindForm(i interface{}) error {
	return formbind.Bind(c.GetForm(), i)
}

func (c *C) Blob(code int, contentType string, data []byte) R {
	c.Log.Infof("status: %d -> %s -> %s", code, contentType, string(data))
	return RFunc(func(c *C, w http.ResponseWriter) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
		_, _ = w.Write(data)
	})
}

func (c *C) Error(status int, err error) R {
	v, _ := json.Marshal(c.MapSet("code", status).Set("msg", err.Error()))
	return c.Blob(200, mimeTypeJSON, v)
}

func (c *C) Errorf(status int, format string, args ...interface{}) R {
	return c.Error(status, fmt.Errorf(format, args...))
}

func (c *C) PrettyJSON(data interface{}, indent string) R {
	v, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return c.Error(500, err)
	}
	return c.Blob(200, mimeTypeJSON, v)
}

func (c *C) JSON(data interface{}) R {
	v, err := json.Marshal(map[string]interface{}{"code": 200, "msg": "OK", "data": data})
	if err != nil {
		return c.Error(500, err)
	}
	return c.Blob(200, mimeTypeJSON, v)
}

func (c *C) JSONArray(data interface{}) R {
	var result []interface{}
	refV := reflect.Indirect(reflect.ValueOf(data))
	switch refV.Type().Kind() {
	case reflect.Slice, reflect.Array:
		for i := 0; i < refV.Len(); i++ {
			result = append(result, extendValue(refV.Index(i)))
		}
	}
	return c.JSON(result)
}

func (c *C) JSONObject(data interface{}) H {
	return c.Map().Extend(data)
}

func (c *C) String(format string, args ...interface{}) R {
	return c.Blob(200, "text/plain; charset=utf-8", []byte(fmt.Sprintf(format, args...)))
}

func (c *C) File(name string) R {
	return RFunc(func(c *C, w http.ResponseWriter) {
		http.ServeFile(w, c.Request, name)
	})
}

func (c *C) Redirect(redirectTo string, permanent bool) R {
	return RFunc(func(c *C, w http.ResponseWriter) {
		code := http.StatusFound
		if permanent {
			code = http.StatusMovedPermanently
		}
		http.Redirect(w, c.Request, redirectTo, code)
	})
}

func (c *C) Write(f func(w http.ResponseWriter)) R {
	return RFunc(func(c *C, w http.ResponseWriter) { f(w) })
}

func (c *C) Map() H {
	return H{}
}

func (c *C) MapSet(name string, value interface{}) H {
	return c.Map().Set(name, value)
}

func (c *C) OK() R {
	return c.MapSet("msg", "OK")
}
