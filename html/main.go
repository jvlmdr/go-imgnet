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
		if _, err := html(root, "", os.Args[2], os.Args[3]); err != nil {
			log.Fatal(err)
		}
	}
}

func html(synset imgnet.Synset, dir, imgnetDir, imgnetURL string) (int, error) {
	log.Printf("%s: %s", synset.WNID, synset.Words)
	// Create a directory with the name of the synset.
	dir = path.Join(dir, synset.WNID)
	if err := os.Mkdir(dir, 0755); err != nil {
		return 0, err
	}

	// Recurse to children.
	counts := make([]int, len(synset.Children))
	for i, child := range synset.Children {
		n, err := html(child, dir, imgnetDir, imgnetURL)
		if err != nil {
			return 0, err
		}
		counts[i] = n
	}

	// Load images in this directory.
	ims, err := imgnet.Images(imgnetDir, synset.WNID)
	if err != nil {
		return 0, err
	}
	// Create index.html.
	if err := saveIndex(path.Join(dir, "index.html"), synset, counts, ims, imgnetURL); err != nil {
		return 0, err
	}

	// Number of images at and below this node.
	count := len(ims)
	for _, n := range counts {
		count += n
	}
	return count, nil
}

func saveIndex(fname string, synset imgnet.Synset, counts []int, ims []string, imgnetURL string) error {
	file, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer file.Close()
	return writeIndex(file, synset, counts, ims, imgnetURL)
}

func writeIndex(w io.Writer, synset imgnet.Synset, counts []int, ims []string, imgnetURL string) error {
	fmt.Fprintf(w, "<h1>%s %s</h1>\n", synset.WNID, synset.Words)
	fmt.Fprintf(w, "<p>%s</p>\n", synset.Gloss)

	// Think of the children.
	fmt.Fprintln(w, "<h2>Children</h2>")
	if len(synset.Children) > 0 {
		fmt.Fprintln(w, "<ul>")
		for i, child := range synset.Children {
			fmt.Fprintf(w, "<li><a href=\"%s\">%s</a> %s (%d)\n", child.WNID, child.WNID, child.Words, counts[i])
		}
		fmt.Fprintln(w, "</ul>")
	} else {
		fmt.Fprintln(w, "<p>None</p>")
	}

	// Print list of images.
	fmt.Fprintf(w, "<h2>Images (%d)</h2>\n", len(ims))
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
