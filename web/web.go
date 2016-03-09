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
	"fmt"
	"html"
	"path/filepath"
	"strings"
	"time"
	"log"

	"honnef.co/go/js/dom"
	"github.com/gopherjs/gopherjs/js"
	"github.com/hanwen/gotiles"
)


func timer(name string) func() {
	start := time.Now()
	return func() { log.Println("timer: ", name, time.Now().Sub(start)) }
}

var service = gotiles.NewService("http://localhost:8080")
var alert *js.Object

func main() {
	js.Global.Set("gotiles", map[string]interface{}{
		"NewUI": NewUI,
	})
	alert = js.Global.Get("alert")
}

func (ui *UI) fetchContent(addr gotiles.TreeAddr) {
	defer timer("fetchContent")()
	c, err := service.BlobContentString(&addr)
	if err != nil {
		ui.contentPane.SetTextContent(html.EscapeString(err.Error()))
		return
	}

	ui.contentPane.SetTextContent("")
	lines := strings.Split(c, "\n")

	base := ui.doc.CreateElement("span")
	base.SetAttribute("class", "content-line")
	for _, l := range lines {
		if len(l) == 0 {
			l = " "
		}
		span := base.CloneNode(false)
		span.SetTextContent(l)
		ui.contentPane.AppendChild(span)
	}

	ui.breadcrumb.SetTextContent(fmt.Sprint("%v", addr))
}

type UI struct {
	addr gotiles.TreeAddr

	doc dom.Document
	breadcrumb dom.Element
	contentPane dom.Element
	fileList  dom.Element
}

func NewUI(repo, branch, dir string) *js.Object {
	doc := dom.GetWindow().Document()
	rv := &UI{
		doc: doc,
		contentPane: doc.GetElementByID("filecontent"),
		fileList: doc.GetElementByID("filelist"),
		breadcrumb: doc.GetElementByID("breadcrumbs"),
	}
	go rv.fetchFileList(gotiles.TreeAddr{
			Repo:   repo,
			Branch: branch,
			Path:    dir,
	})
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

func (ui *UI) onFileListClickTree(e dom.Event) {
	e.StopPropagation()

	// todo - don't go through DOM
	name := e.Target().TextContent()

	addr := ui.addr
	if addr.Path == "" {
		addr.Path = "."
	}
	addr.Path = strings.TrimLeft(filepath.Join(addr.Path, name), "/")
	go ui.fetchFileList(addr)
}

func (ui *UI) fetchFileList(addr gotiles.TreeAddr) {
	defer timer("fetchFileList")()
	t, err := service.Tree(&addr)
	if err != nil {
		alert.Invoke(html.EscapeString(err.Error()))
		return
	}

	if !(addr.Path == "" || addr.Path == ".") {
		t.Entries = append(t.Entries,
			gotiles.TreeEntry{
				Name: "..",
				Mode: 040000,
				Type: "tree",
			})
	}

	ui.fileList.SetTextContent("")
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
		} else if t == "tree" {
			span.AddEventListener("click", true, ui.onFileListClickTree)
		}

		ui.fileList.AppendChild(span)
	}

	ui.addr = addr
}
