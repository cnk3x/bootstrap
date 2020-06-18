package dig

import (
	"go.uber.org/fx"
)

var containers []fx.Option

func Add(opts ...fx.Option) {
	containers = append(containers, opts...)
}

func Provide(constructors ...interface{}) {
	Add(fx.Provide(constructors...))
}

func Invoke(constructors ...interface{}) {
	Add(fx.Invoke(constructors...))
}

func Supply(values ...interface{}) {
	Add(fx.Supply(values...))
}

func Populate(targets ...interface{}) {
	Add(fx.Populate(targets...))
}

func Run() {
	App().Run()
}

func App() *fx.App {
	return fx.New(containers...)
}
