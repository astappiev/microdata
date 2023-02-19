package microdata

import (
	"bytes"
	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
	"io"
	"net/http"
	"net/url"
)

// ParseURL parses the HTML document available at the given URL and returns the microdata.
func ParseURL(urlStr string) (*Microdata, error) {
	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		return nil, err
	}

	contentType := resp.Header.Get("Content-Type")
	return ParseHTML(resp.Body, contentType, urlStr)
}

// ParseHTML parses the HTML document available in the given reader and returns the microdata. The given url is
// used to resolve the URLs in the attributes. The given contentType is used to convert the content of r to UTF-8.
// When the given contentType is equal to "", the content type will be detected using `http.DetectContentType`.
func ParseHTML(r io.Reader, contentType string, urlStr string) (*Microdata, error) {
	if contentType == "" {
		b := make([]byte, 512)
		_, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		contentType = http.DetectContentType(b)
		r = io.MultiReader(bytes.NewReader(b), r)
	}

	r, err := charset.NewReader(r, contentType)
	if err != nil {
		return nil, err
	}

	tree, err := html.Parse(r)
	if err != nil {
		return nil, err
	}

	return ParseNode(tree, urlStr)
}

// ParseNode parses the root Node and returns the microdata.
func ParseNode(root *html.Node, urlStr string) (*Microdata, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	p, err := newParser(root, u)
	if err != nil {
		return nil, err
	}

	return p.parse()
}
