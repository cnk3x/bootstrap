package mux

import (
    "net/http"
)

var _ R = RFunc(nil)

type R interface {
    WriteTo(c *C, w http.ResponseWriter)
}

type RFunc func(c *C, w http.ResponseWriter)

func (f RFunc) WriteTo(c *C, w http.ResponseWriter) {
    f(c, w)
}
