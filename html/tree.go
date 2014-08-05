package main

import (
	"os"

	"github.com/jvlmdr/go-imgnet/imgnet"
)

func loadTree(fname string) (imgnet.Tree, error) {
	file, err := os.Open(fname)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	return imgnet.DecodeTree(file)
}
