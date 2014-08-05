package imgnet

import (
	"encoding/xml"
	"io"
)

// Tree is represented by a list of root nodes.
type Tree []Synset

type Synset struct {
	WNID     string   `xml:"wnid,attr"`
	Words    string   `xml:"words,attr"`
	Gloss    string   `xml:"gloss,attr"`
	Children []Synset `xml:"synset"`
}

type root struct {
	XMLName  xml.Name `xml:"ImageNetStructure"`
	Children []Synset `xml:"synset"`
}

func DecodeTree(r io.Reader) (Tree, error) {
	var x root
	if err := xml.NewDecoder(r).Decode(&x); err != nil {
		return nil, err
	}
	return x.Children, nil
}
