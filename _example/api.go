package main

import (
	"go.shu.run/bootstrap/dig"
	"go.shu.run/bootstrap/mux"
)

func init() {
	dig.Invoke(Api)
}

func Api(router *mux.Mux) {
	router.GET("/", func(c *mux.C) mux.R {
		return mux.H{"msg": "OK"}
	})
}
