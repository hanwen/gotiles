// +build !js

package gotiles
import (
	"encoding/base64"
)

func nativeAtoB(in []byte) string {
	c, err := base64.StdEncoding.DecodeString(string(in))
	if err != nil {
		return "error: " + err.Error()
	}
	return string(c)
}
