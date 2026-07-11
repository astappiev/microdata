package microdata

import (
	"golang.org/x/net/html"
)

// getAttr returns the value associated with the given attribute from the given node.
func getAttr(attribute string, node *html.Node) (string, bool) {
	for _, attr := range node.Attr {
		if attr.Key == attribute {
			return attr.Val, true
		}
	}
	return "", false
}

// walkNodes traverses the node tree executing the given functions.
func walkNodes(root *html.Node, f func(*html.Node)) {
	if root == nil {
		return
	}
	f(root)
	for n := range root.Descendants() {
		f(n)
	}
}
