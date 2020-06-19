package mux

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"time"

	"go.shu.run/bootstrap/logger"
	"go.shu.run/bootstrap/mux/binding"
)

const (
	modelKey     = "__model__"
	pathParamKey = "__path_param__"

	mimeTypeJSON = "application/json; charset=utf-8"
)

var _ context.Context = (*C)(nil)

//C mux's Context
type C struct {
	mux *Mux
	log logger.Logger

	values      map[string]interface{} //中间件值载体
	respHeaders http.Header            //输出的http头

	*http.Request

	sameSite http.SameSite
	cookies  []*http.Cookie
}

//impl context

//Deadline Request's deadline
func (c *C) Deadline() (deadline time.Time, ok bool) {
	return c.Request.Context().Deadline()
}

//Done request's Done
func (c *C) Done() <-chan struct{} {
	return c.Request.Context().Done()
}

//Err request's contxt's Err
func (c *C) Err() error {
	return c.Request.Context().Err()
}

//Value request's context's values
func (c *C) Value(key interface{}) interface{} {
	return c.Request.Context().Value(key)
}

func (c *C) reset() {
	c.Request = nil
	if len(c.values) > 0 {
		for key := range c.values {
			delete(c.values, key)
		}
	}

	for n := range c.respHeaders {
		c.respHeaders.Del(n)
	}
}

func (c *C) release() {
	c.reset()
	c.mux.cPool.Put(c)
}

//Set set
func (c *C) Set(name string, value interface{}) {
	c.values[name] = value
}

//Get get
func (c *C) Get(name string) interface{} {
	if len(c.values) > 0 {
		return c.values[name]
	}
	return nil
}

// input

//Param param
func (c *C) Param(name string) string {
	v, ok := c.Get(pathParamKey + name).(string)
	if !ok {
		return "-"
	}
	return v
}

//ParamInt get int from params
func (c *C) ParamInt(name string) int64 {
	v, _ := strconv.ParseInt(c.Param(name), 10, 64)
	return v
}

//FormValue get string from form
func (c *C) FormValue(names ...string) string {
	for _, name := range names {
		s := c.Request.FormValue(name)
		if s != "" {
			return s
		}
	}
	return ""
}

//FormInt get int from form
func (c *C) FormInt(names ...string) int64 {
	v, _ := strconv.ParseInt(c.FormValue(names...), 10, 64)
	return v
}

//FormFloat get float from form
func (c *C) FormFloat(names ...string) float64 {
	v, _ := strconv.ParseFloat(c.FormValue(names...), 64)
	return v
}

//FormBool get bool from Form
func (c *C) FormBool(names ...string) bool {
	v, _ := strconv.ParseBool(c.FormValue(names...))
	return v
}

//SetModel set model value
func (c *C) SetModel(value interface{}) {
	c.Set(modelKey, value)
}

//Model get model value
func (c *C) Model() interface{} {
	return c.Get(modelKey)
}

//GetHostURL get fully url
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

//GetPageURI getf fully url
func (c *C) GetPageURI() string {
	if c.URL.IsAbs() {
		return c.URL.String()
	}
	return c.GetHostURL() + c.URL.String()
}

//GetURI ...
func (c *C) GetURI(format string, args ...interface{}) string {
	return c.GetHostURL() + fmt.Sprintf(format, args...)
}

//GetForm form
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

//Bind bind
func (c *C) Bind(i interface{}) error {
	binder := binding.Default(c.Method, c.HeaderGet("Content-Type"))
	return binder.Bind(c.Request, i)
}

//func (c *C) BindForm(i interface{}) error {
//	return formbind.Bind(c.GetForm(), i)
//}

//header

//HeaderGet HeaderGet
func (c *C) HeaderGet(names ...string) string {
	for _, name := range names {
		s := c.Request.Header.Get(name)
		if s != "" {
			return s
		}
	}
	return ""
}

//GetCookie GetCookie
func (c *C) GetCookie(name string) string {
	cookie, err := c.Request.Cookie(name)
	if err != nil {
		return ""
	}
	val, _ := url.QueryUnescape(cookie.Value)
	return val
}

// output

// header

//HeaderSet HeaderSet
func (c *C) HeaderSet(name, value string) {
	c.respHeaders.Set(name, value)
}

//HeaderAdd HeaderAdd
func (c *C) HeaderAdd(name, value string) {
	c.respHeaders.Add(name, value)
}

// SetSameSite with cookie
func (c *C) SetSameSite(sameSite http.SameSite) {
	c.sameSite = sameSite
}

// SetCookie adds a Set-Cookie header to the ResponseWriter's headers.
// The provided cookie must have a valid Name.
// Invalid cookies may be silently dropped.
func (c *C) SetCookie(name, value string, maxAge int, path, domain string, secure, httpOnly bool) {
	if path == "" {
		path = "/"
	}

	c.cookies = append(c.cookies, &http.Cookie{
		Name:     name,
		Value:    url.QueryEscape(value),
		MaxAge:   maxAge,
		Path:     path,
		Domain:   domain,
		SameSite: c.sameSite,
		Secure:   secure,
		HttpOnly: httpOnly,
	})
}

//response

//Blob Blob
func (c *C) Blob(code int, contentType string, data []byte) R {
	c.log.Infof("status: %d -> %s -> %s", code, contentType, string(data))
	return RFunc(func(c *C, w http.ResponseWriter) {
		w.Header().Set("Content-Type", contentType)
		w.WriteHeader(code)
		_, _ = w.Write(data)
	})
}

//Error Error
func (c *C) Error(status int, err error) R {
	v, _ := json.Marshal(c.MapSet("code", status).Set("msg", err.Error()))
	return c.Blob(200, mimeTypeJSON, v)
}

//Errorf Errorf
func (c *C) Errorf(status int, format string, args ...interface{}) R {
	return c.Error(status, fmt.Errorf(format, args...))
}

//PrettyJSON PrettyJSON
func (c *C) PrettyJSON(data interface{}, indent string) R {
	v, err := json.MarshalIndent(data, "", indent)
	if err != nil {
		return c.Error(500, err)
	}
	return c.Blob(200, mimeTypeJSON, v)
}

//JSON JSON
func (c *C) JSON(data interface{}) R {
	v, err := json.Marshal(data)
	if err != nil {
		return c.Error(500, err)
	}
	return c.Blob(200, mimeTypeJSON, v)
}

//JSONArray JSONArray
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

//JSONObject JSONObject
func (c *C) JSONObject(data interface{}) H {
	return c.Map().Extend(data)
}

//String String
func (c *C) String(format string, args ...interface{}) R {
	return c.Blob(200, "text/plain; charset=utf-8", []byte(fmt.Sprintf(format, args...)))
}

//File File
func (c *C) File(name string) R {
	return RFunc(func(c *C, w http.ResponseWriter) {
		http.ServeFile(w, c.Request, name)
	})
}

//Redirect Redirect
func (c *C) Redirect(redirectTo string, permanent bool) R {
	return RFunc(func(c *C, w http.ResponseWriter) {
		code := http.StatusFound
		if permanent {
			code = http.StatusMovedPermanently
		}
		http.Redirect(w, c.Request, redirectTo, code)
	})
}

//Write Write
func (c *C) Write(f func(w http.ResponseWriter)) R {
	return RFunc(func(c *C, w http.ResponseWriter) { f(w) })
}

//Map Map
func (c *C) Map() H {
	return H{}
}

//MapSet MapSet
func (c *C) MapSet(name string, value interface{}) H {
	return c.Map().Set(name, value)
}

//OK OK
func (c *C) OK() R {
	return c.MapSet("msg", "OK")
}
