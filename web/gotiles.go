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

package main

import (
	"bytes"
	"html"
	"path/filepath"
	"strings"

	"honnef.co/go/js/dom"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanwen/gotiles"
)


var service = gotiles.NewService("http://localhost:8080")
var alert *js.Object

func main() {
	js.Global.Set("gotiles", map[string]interface{}{
		"NewUI": NewUI,
	})
	alert = js.Global.Get("alert")
}

func (ui *UI) fetchContent(addr gotiles.TreeAddr) {
	c, err := service.BlobContent(&addr)
	if err != nil {
		ui.contentPane.SetTextContent(html.EscapeString(err.Error()))
		return
	}

	ui.contentPane.SetTextContent("")
	lines := bytes.Split(c, []byte("\n"))
	for _, l := range lines {
		span := ui.doc.CreateElement("span")
		span.SetAttribute("class", "content-line")
		span.SetTextContent(string(l))
		ui.contentPane.AppendChild(span)
	}
}


type UI struct {
	addr gotiles.TreeAddr

	doc dom.Document
	contentPane dom.Element
	fileList  dom.Element
}

func NewUI(repo, branch, dir string) *js.Object {
	doc := dom.GetWindow().Document()
	rv := &UI{
		doc: doc,
		addr: gotiles.TreeAddr{
			Repo:   repo,
			Branch: branch,
			Path:    dir,
		},
		contentPane: doc.GetElementByID("filecontent"),
		fileList: doc.GetElementByID("filelist"),
	}
	go rv.fetchFileList()
	return js.MakeWrapper(rv)
}

func (ui *UI) onFileListClick(e dom.Event) {
	e.StopPropagation()

	// todo - don't go through DOM
	name := e.Target().TextContent()

	addr := ui.addr
	if addr.Path == "" {
		addr.Path = "."
	}
	addr.Path = strings.TrimLeft(filepath.Join(addr.Path, name), "/")
	go ui.fetchContent(addr)
}

func (ui *UI) fetchFileList() {
	t, err := service.Tree(&ui.addr)
	if err != nil {
		alert.Invoke(html.EscapeString(err.Error()))
		return
	}

	for _, e := range t.Entries {
		t := e.Type
		switch e.Mode {
		case 0100755:
			t = "xblob"
		case 0120000:
			// if only the json gave us the link target as well....
			t = "slink"
		}
		_ = t
		span := ui.doc.CreateElement("span")
		span.SetAttribute("class", "fileentry")
		span.SetTextContent(e.Name)
		if t == "blob" || t == "xblob" {
			span.AddEventListener("click", true, ui.onFileListClick)
		}
		ui.fileList.AppendChild(span)
	}
}
