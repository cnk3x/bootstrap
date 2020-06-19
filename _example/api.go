package main

import (
	"go.shu.run/bootstrap/mux"
)

func api(router *mux.Mux) {
	router.GET("/", func(c *mux.C) mux.R {
		return mux.H{"msg": "OK"}
	})
	router.GET("/ping", func(c *mux.C) mux.R {
		return mux.H{"msg": "TONG"}
	})
}
