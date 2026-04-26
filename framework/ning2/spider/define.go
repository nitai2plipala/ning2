package spider

import (
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
	"net/http"
)

type Engine struct {
	HttpClient *http.Client
	Header map[string]string
	BodyCode    string
}

type HtmNode struct {
	Parent, FirstChild, LastChild, PrevSibling, NextSibling *HtmNode

	Type      html.NodeType
	DataAtom  atom.Atom
	Data      string
	Namespace string
	Attr      []html.Attribute
}
