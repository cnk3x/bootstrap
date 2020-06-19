package mux

import (
	"net/http"
)

var _ R = RFunc(nil)

//R response interface
type R interface {
	WriteTo(c *C, w http.ResponseWriter)
}

//RFunc response function
type RFunc func(c *C, w http.ResponseWriter)

//WriteTo response
func (f RFunc) WriteTo(c *C, w http.ResponseWriter) {
	f(c, w)
}
