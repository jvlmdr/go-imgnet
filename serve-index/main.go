package main

import (
	"bytes"
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-imgnet"
	"html/template"
	"io"
	"net/http"
	"os"
	"path"
	"regexp"
	"strconv"
)

func main() {
	flag.Parse()

	if flag.NArg() != 3 {
		binary := path.Base(os.Args[0])
		fmt.Printf("Usage: %s dir index port\n", binary)
		return
	}

	filename := flag.Arg(1)
	port, err := strconv.Atoi(flag.Arg(2))
	if err != nil {
		fmt.Println(err)
		return
	}

	// Load index
	var index imgnet.Index
	file, err := os.Open(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	dec := gob.NewDecoder(file)
	err = dec.Decode(&index)
	if err != nil {
		fmt.Println(err)
		return
	}

	http.HandleFunc("/browse/", func(w http.ResponseWriter, r *http.Request) {
		HandleBrowse(w, r, index)
	})

	addr := fmt.Sprintf(":%d", port)
	http.ListenAndServe(addr, nil)
}

//
//
//
func HandleBrowse(w http.ResponseWriter, r *http.Request, index imgnet.Index) {
	path := r.URL.Path

	if matched, _ := regexp.MatchString(`^/browse/$`, path); matched {
		HandleBrowseRoot(w, index)
		return
	}

	// Need trailing slash so that URLs are treated as relative
	re := regexp.MustCompile(`^/browse/([^/]+)/$`)
	matches := re.FindStringSubmatch(path)
	if len(matches) > 0 {
		HandleBrowseSynset(w, r, index, matches[1])
		return
	}

	// No trailing slash since this is like a file in a folder
	re = regexp.MustCompile(`^/browse/([^/]+)/([^/]+)$`)
	matches = re.FindStringSubmatch(path)
	if len(matches) > 0 {
		HandleBrowseImage(w, index, matches[1], matches[2])
		return
	}

	http.Error(w, "Invalid URL", http.StatusNotFound)
}

//
//
//
func HandleBrowseRoot(w http.ResponseWriter, index imgnet.Index) {
	// Load template
	tmpl, err := template.ParseFiles("browse.html")
	if err != nil {
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		return
	}

	// Execute template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, index)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

//
//
//
func HandleBrowseSynset(w http.ResponseWriter, r *http.Request, index imgnet.Index, synset string) {
	// Check that synset exists
	_, present := index.Synsets[synset]
	if !present {
		http.Error(w, "Could not find synset", http.StatusNotFound)
		return
	}

	images, err := index.SynsetIndex(synset)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not load synset index", http.StatusInternalServerError)
		return
	}

	// Load template
	tmpl, err := template.ParseFiles("browse-synset.html")
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not parse template", http.StatusInternalServerError)
		return
	}

	// Execute template
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, images)
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not execute template", http.StatusInternalServerError)
		return
	}
	io.Copy(w, &buf)
}

//
//
//
func HandleBrowseImage(w http.ResponseWriter, index imgnet.Index, synset, name string) {
	image, err := index.Open(imgnet.Image{synset, name})
	if err != nil {
		fmt.Println(err)
		http.Error(w, "Could not open image", http.StatusNotFound)
	}

	io.Copy(w, image)
}
