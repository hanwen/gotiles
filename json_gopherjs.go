// Copyright 2016 Google Inc. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// +build !linux
package gotiles

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gopherjs/gopherjs/js"
)

var jsonObj = js.Global.Get("JSON")
var _ = log.Println

func jsonUnmarshal(content []byte, dest interface{}) error {
	return json.Unmarshal(content, dest)

	// TODO
	// * benchmark against encoding/json,
	// * reflection?
	obj := jsonObj.Call("parse", string(content))
	return jsUnmarshal(obj, dest)
}

func jsUnmarshal(o *js.Object, dest interface{}) error {
	switch t := dest.(type) {
	case *TreeResponse:
		t.ID = o.Get("id").String()
		entries := o.Get("entries")
		l := entries.Length()
		for i := 0; i < l; i++ {
			var e TreeEntry
			if err := jsUnmarshal(entries.Index(i), &e); err != nil {
				return err
			}
			t.Entries = append(t.Entries, e)
		}
	case *TreeEntry:
		t.Mode = o.Get("mode").Int()
		t.Type = o.Get("type").String()
		t.ID = o.Get("id").String()
		t.Name = o.Get("name").String()

	default:
		return fmt.Errorf("unknown type %T", dest)
	}
	return nil
}
