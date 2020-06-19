package main

import (
	"go.shu.run/bootstrap"
	"go.shu.run/bootstrap/dig"
)

func main() {
	bootstrap.AddHTTPServer()

	dig.Invoke(api)

	bootstrap.Run()
}
