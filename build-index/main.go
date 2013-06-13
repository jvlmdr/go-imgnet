package main

import (
	"encoding/gob"
	"flag"
	"fmt"
	"github.com/jackvalmadre/go-imgnet"
	"os"
	"path"
)

func main() {
	verbose := flag.Bool("verbose", false, "Print synsets as they come")
	flag.Parse()
	if flag.NArg() != 2 {
		binary := path.Base(os.Args[0])
		fmt.Printf("Usage: %s [options] dir index\n", binary)
		return
	}
	dir := flag.Arg(0)
	filename := flag.Arg(1)

	// Construct index.
	index, err := imgnet.BuildIndex(dir, *verbose)
	if err != nil {
		fmt.Println(err)
		return
	}

	n := 0
	for _, v := range index.Synsets {
		n += v
	}

	fmt.Println("Index contains", n, "images between", len(index.Synsets), "synsets")

	// Save to file.
	file, err := os.Create(filename)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer file.Close()

	enc := gob.NewEncoder(file)
	err = enc.Encode(index)
	if err != nil {
		fmt.Println(err)
		return
	}
}
