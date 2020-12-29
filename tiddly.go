// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"crypto/md5"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
)

func init() {
	http.HandleFunc("/", index)
	http.HandleFunc("/status", status)
	http.HandleFunc("/recipes/all/tiddlers/", tiddler)
	http.HandleFunc("/recipes/all/tiddlers.json", tiddlerList)
	http.HandleFunc("/bags/bag/tiddlers/", deleteTiddler)
}

type Tiddler struct {
	Rev  int
	Meta string
	Text string
}

func index(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "bad method", 405)
		return
	}
	if r.URL.Path != "/" {
		http.Error(w, "not found", 404)
		return
	}

	http.ServeFile(w, r, filepath.Join(tiddlerFolder, "index.html"))
}

func status(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	name := "GUEST"
	w.Write([]byte(`{"username": "` + name + `", "space": {"recipe": "all"}}`))
}

func tiddlerList(w http.ResponseWriter, r *http.Request) {
	var buf bytes.Buffer
	sep := ""
	buf.WriteString("[")

	err := filepath.Walk(tiddlerFolder, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			var t Tiddler
			key, _ := filepath.Rel(tiddlerFolder, path)
			err = fsget(key, &t)
			if err != nil {
				return nil
			}

			if len(t.Meta) == 0 {
				return nil
			}

			meta := t.Meta

			// Tiddlers containing macros don't take effect until
			// they are loaded. Force them to be loaded by including
			// their bodies in the skinny tiddler list.
			// Might need to expand this to other kinds of tiddlers
			// in the future as we discover them.
			if strings.Contains(meta, `"$:/tags/Macro"`) {
				var js map[string]interface{}
				uerr := json.Unmarshal([]byte(meta), &js)
				if uerr != nil {
					return nil
				}
				js["text"] = string(t.Text)
				data, eerr := json.Marshal(js)
				if eerr != nil {
					return nil
				}
				meta = string(data)
			}

			buf.WriteString(sep)
			sep = ","
			buf.WriteString(meta)

		}
		return nil
	})
	if err != nil {
		return
	}

	buf.WriteString("]")
	w.Header().Set("Content-Type", "application/json")
	w.Write(buf.Bytes())
}

func tiddler(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case "GET":
		getTiddler(w, r)
	case "PUT":
		putTiddler(w, r)
	default:
		http.Error(w, "bad method", 405)
	}
}

func fskey(title string) string {
	sum := sha256.Sum256([]byte(title))
	key := base64.URLEncoding.EncodeToString(sum[:])
	return key
}

func fsget(key string, t *Tiddler) error {
	ifile, ierr := os.Open(filepath.Join(tiddlerFolder, key))
	if ierr != nil {
		return ierr
	}
	jdec := json.NewDecoder(ifile)
	jerr := jdec.Decode(&t)
	if jerr != nil {
		return jerr
	}
	return ifile.Close()
}

func fsput(key string, t *Tiddler) error {
	opath := filepath.Join(tiddlerFolder, key)
	odir := filepath.Dir(opath)
	os.MkdirAll(odir, 0770)

	ofile, oerr := os.Create(opath)
	if oerr != nil {
		return oerr
	}
	jdec := json.NewEncoder(ofile)
	jerr := jdec.Encode(&t)
	if jerr != nil {
		return jerr
	}
	return ofile.Close()
}

func getTiddler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/recipes/all/tiddlers/")

	var t Tiddler
	key := fskey(title)
	if err := fsget(key, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	var js map[string]interface{}
	err := json.Unmarshal([]byte(t.Meta), &js)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	js["text"] = string(t.Text)
	data, err := json.Marshal(js)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func putTiddler(w http.ResponseWriter, r *http.Request) {
	title := strings.TrimPrefix(r.URL.Path, "/recipes/all/tiddlers/")
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "cannot read data", 400)
		return
	}
	var js map[string]interface{}
	err = json.Unmarshal(data, &js)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	js["bag"] = "bag"

	rev := 1
	var old Tiddler
	key := fskey(title)
	if err := fsget(key, &old); err == nil {
		rev = old.Rev + 1
	}
	js["revision"] = rev

	var t Tiddler
	text, ok := js["text"].(string)
	if ok {
		t.Text = text
	}
	delete(js, "text")
	t.Rev = rev
	meta, err := json.Marshal(js)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	t.Meta = string(meta)
	err = fsput(key, &t)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	etag := fmt.Sprintf("\"bag/%s/%d:%x\"", url.QueryEscape(title), rev, md5.Sum(data))
	w.Header().Set("Etag", etag)
}

func deleteTiddler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "bad method", 405)
		return
	}
	title := strings.TrimPrefix(r.URL.Path, "/bags/bag/tiddlers/")
	var t Tiddler
	key := fskey(title)
	if err := fsget(key, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	t.Rev++
	t.Meta = ""
	t.Text = ""
	if err := fsput(key, &t); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
