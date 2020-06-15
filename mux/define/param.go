package define

import "fmt"

type Param struct {
    name     string      //名称
    desc     string      //描述
    require  bool        //必须
    message  string      //
    validate []Validator //格式验证
}

func (p *Param) Add(vs ...Validator) *Param {
    p.validate = append(p.validate, vs...)
    return p
}

func (p *Param) Require() *Param {
    return p.RequireMessage(fmt.Sprintf("缺少参数: %s(%s):", p.desc, p.name))
}

func (p *Param) RequireMessage(msg string) *Param {
    p.require = true
    p.message = msg
    return p
}
