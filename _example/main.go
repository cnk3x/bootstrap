package main

import (
	"go.shu.run/bootstrap"
	"go.shu.run/bootstrap/dig"
)

func main() {
	dig.Invoke(bootstrap.StartMux)
	bootstrap.Run()
}
