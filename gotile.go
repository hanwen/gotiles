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

// Package gotiles is a client library for the Gitiles source viewer.
package gotiles

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Service struct {
	addr string
}

func NewService(addr string) *Service {
	return &Service{addr: addr}
}

type TreeEntry struct {
	Mode int
	Type string
	ID   string
	Name string
}

type TreeResponse struct {
	ID      string
	Entries []TreeEntry
}

type TreeOrErrorResponse struct {
	*TreeResponse
	Error error
}

func (s *Service) BlobContentString(addr *TreeAddr) (string, error) {
	url := fmt.Sprintf("%s/%s/+show/%s/%s?format=TEXT", s.addr, addr.Repo, addr.Branch, addr.Path)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	c, err :=  ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	out := nativeAtoB(c)
	return out, err
}

func (s *Service) Tree(addr *TreeAddr) (*TreeResponse, error) {
	dir := addr.Path
	if dir == "" {
		dir = "."
	}
	url := fmt.Sprintf("%s/%s/+/%s/%s?format=JSON", s.addr, addr.Repo, addr.Branch, dir)
	var entry TreeResponse
	err := getJSON(url, &entry)
	return &entry, err
}

func getJSON(url string, dest interface{}) error {
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	c, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	i := bytes.IndexByte(c, '\n')
	if i < 0 {
		return fmt.Errorf("missing header, %q", c)
	}

	c = c[i+1:]

	err = unmarshalGitiles(c, dest)
	return err
}

func unmarshalGitiles(c []byte, dest interface{}) error {
	if err := jsonUnmarshal(c, dest); err != nil {
		return err
	}

	return nil
}

type TreeAddr struct {
	Repo   string
	Branch string
	Path   string
}

type Project struct {
	Name     string
	CloneURL string `json:"clone_url"`
}

type Person struct {
	Name  string
	Email string
	Time  string // TODO - time.Time.
}

type DiffEntry struct {
	Type    string
	OldID   string `json:"old_id"`
	OldMode int    `json:"old_mode"`
	OldPath string `json:"old_path"`
	NewID   string `json:"new_id"`
	NewMode int    `json:"new_mode"`
	NewPath string `json:"new_path"`
}

type Commit struct {
	Commit    string
	Tree      string
	Parents   []string
	Author    Person
	Committer Person
	Message   string
	TreeDiff  []DiffEntry `json:"tree_diff"`
}
type Log struct {
	Log  []Commit
	Next string
}

type BlameRegion struct {
	Start  int
	Count  int
	Path   string
	Commit string
	Author Person
}

type Blame struct {
	Regions []BlameRegion
}
