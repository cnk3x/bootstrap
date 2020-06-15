package define

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

func New() *ApiDefine {
	return &ApiDefine{
		inMap:  make(map[string]*Param),
		outMap: make(map[string]*Param),
	}
}

type ApiDefine struct {
	in     []string
	inMap  map[string]*Param
	out    []string
	outMap map[string]*Param
}

func (d *ApiDefine) In(name string, desc string) *Param {
	param := &Param{name: name, desc: desc}
	if d.inMap == nil {
		d.inMap = make(map[string]*Param, 1)
	}
	if _, find := d.inMap[name]; !find {
		d.in = append(d.in, name)
	}
	d.inMap[name] = param

	return param
}

func (d *ApiDefine) Out(name string, desc string) *Param {
	param := &Param{name: name, desc: desc}
	if d.outMap == nil {
		d.outMap = make(map[string]*Param, 1)
	}
	if _, find := d.outMap[name]; !find {
		d.out = append(d.out, name)
	}
	d.outMap[name] = param

	return param
}

func (d *ApiDefine) BaseOut() {
	d.Out("code", "返回代码")
	d.Out("msg", "返回说明")
}

func (d *ApiDefine) Validate(u url.Values) error {
	errs := Errors{}
	for _, name := range d.in {
		p := d.inMap[name]
		value := u.Get(name)

		if value == "" {
			if p.require {
				errs.Add(name, p.message)
			}
			continue
		}

		for _, v := range p.validate {
			if err := v.Exec(value); err != nil {
				errs.Add(name, err.Error())
			}
		}
	}
	if errs.IsEmpty() {
		return nil
	}
	return errs
}

func (d *ApiDefine) validateRequest(req *http.Request) error {
	var form url.Values
	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if err := req.ParseForm(); err != nil {
			return err
		}
		form = req.Form
	case http.MethodGet:
		form = req.URL.Query()
	default:
		return fmt.Errorf("不支持的方法: %s", req.Method)
	}

	return d.Validate(form)
}

func (d *ApiDefine) ValidateRequest(req *http.Request) func(w http.ResponseWriter) {
	var form url.Values
	switch req.Method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		if err := req.ParseForm(); err != nil {
			return errWriter(err)
		}
		form = req.Form
	case http.MethodGet:
		form = req.URL.Query()
	default:
		return errWriter(fmt.Errorf("不支持的方法: %s", req.Method))
	}

	if err := d.Validate(form); err != nil {
		switch ex := err.(type) {
		case Errors:
			return jsonWriter(400, "args", ex)
		default:
			return errWriter(err)
		}
	}

	return nil
}

func errWriter(err error) func(w http.ResponseWriter) {
	return jsonWriter(400, "err", err.Error())
}

func jsonWriter(code int, name string, value interface{}) func(w http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		data, _ := json.Marshal(map[string]interface{}{name: value})
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(code)
		_, _ = w.Write(data)
	}
}
