// +build js

package gotiles

import (
	"github.com/gopherjs/gopherjs/js"
)

var base64Decode *js.Object
func init() {
	base64Decode = js.Global.Get("atob")
}


func nativeAtoB(in []byte) string {
	// TODO - the bytes/string conversion actually takes a big
	// chunk of time.
	result := base64Decode.Invoke(string(in))
	return result.String()
}
