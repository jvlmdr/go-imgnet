package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"

	"github.com/jvlmdr/go-imgnet/imgnet"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Fprintf(os.Stderr, "usage: %s tree.xml imgnet-dir imgnet-url\n", os.Args[0])
		os.Exit(1)
	}

	tree, err := loadTree(os.Args[1])
	if err != nil {
		log.Fatal(err)
	}

	for _, root := range tree {
		if err := html(root, "", os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	}
}

func html(synset imgnet.Synset, dir, imgnetDir, imgnetURL string) error {
	log.Printf("%s: %s", synset.WNID, synset.Words)
	// Create a directory with the name of the synset.
	dir = path.Join(dir, synset.WNID)
	if err := os.Mkdir(dir, 0755); err != nil {
		return err
	}
	// Create index.html.
	if err := saveIndex(synset, path.Join(dir, "index.html"), imgnetDir, imgnetURL); err != nil {
		return err
	}
	// Recurse to children.
	for _, child := range synset.Children {
		if err := html(child, dir, imgnetDir, imgnetURL); err != nil {
			return err
		}
	}
	return nil
}

func saveIndex(synset imgnet.Synset, fname, imgnetDir, imgnetURL string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeIndex(synset, file, imgnetDir, imgnetURL)
}

func writeIndex(synset imgnet.Synset, w io.Writer, imgnetDir, imgnetURL string) error {
	ims, err := imgnet.Images(imgnetDir, synset.WNID)
	if err != nil {
		return err
	}

	fmt.Fprintf(w, "<h1>%s %s</h1>\n", synset.WNID, synset.Words)
	fmt.Fprintf(w, "<p>%s</p>\n", synset.Gloss)

	// Think of the children.
	fmt.Fprintln(w, "<h2>Children</h2>")
	if len(synset.Children) > 0 {
		fmt.Fprintln(w, "<ul>")
		for _, child := range synset.Children {
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a> %s\n", child.WNID, child.WNID, child.Words)
		}
		fmt.Fprintln(w, "</ul>")
	} else {
		fmt.Fprintln(w, "<p>None</p>")
	}

	// Print list of images.
	fmt.Fprintln(w, "<h2>Images</h2>")
	if len(ims) > 0 {
		fmt.Fprintln(w, "<ul>")
		for _, im := range ims {
			url := path.Join(imgnetURL, synset.WNID, im)
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a>\n", url, im)
		}
		fmt.Fprintln(w, "</ul>")
	} else {
		fmt.Fprintln(w, "<p>None</p>")
	}
	return nil
}
