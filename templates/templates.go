package templates

import (
	"io/ioutil"
	"net/http"
)

type Template struct {
	fs http.FileSystem
}

func (tpl *Template) readBytes(name string) ([]byte, error) {
	f, err := tpl.fs.Open(name)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func (tpl *Template) readDefine(name string) ([]byte, error) {
	f, err := tpl.fs.Open(name + ".json")
	if err != nil {
		return nil, err
	}
	return ioutil.ReadAll(f)
}

func (tpl *Template) Render() {

}

type routerData struct {
	Layout  string                 `json:"layout"`  //布局文件路径, 相对当前文件
	Data    map[string]interface{} `json:"data"`    //本地数据定义
	Load    map[string]string      `json:"load"`    //远程数据
	Include []string               `json:"include"` //包含扩展的定义
}
