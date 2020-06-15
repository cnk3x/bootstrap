package formbind

import (
	"encoding/json"
	"net/url"
	"testing"
)

type Obj struct {
	Name     string
	FullName []string
	Age      int
	Height   *float32
	Weight   float64
	Fav      *[]string
}

func TestFormBind(t *testing.T) {
	var q = `name=shu&age=38&height=171&fav=game&fav=ml&fav=sleep&full_name=shu&full_name=s&full_name=x`
	u, _ := url.ParseQuery(q)
	var obj Obj
	if err := Bind(u, &obj); err != nil {
		t.Errorf("%v", err)
	}

	v, _ := json.MarshalIndent(obj, "", "  ")
	t.Logf(string(v))
}
