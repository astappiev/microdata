package microdata

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"net/url"

	"golang.org/x/net/html"
	"golang.org/x/net/html/charset"
)

// ParseURL parses the HTML document available at the given URL and returns the microdata.
func ParseURL(urlStr string) (*Microdata, error) {
	return ParseURLWithContext(context.Background(), urlStr)
}

// ParseURLWithContext parses the HTML document available at the given URL using the provided context and returns the microdata.
func ParseURLWithContext(ctx context.Context, urlStr string) (*Microdata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	contentType := resp.Header.Get("Content-Type")
	return ParseHTML(resp.Body, contentType, resp.Request.URL.String())
}

// ParseHTML parses the HTML document available in the given reader and returns the microdata. The given url is
// used to resolve the URLs in the attributes. The given contentType is used to convert the content of r to UTF-8.
// When the given contentType is equal to "", the content type will be detected using `http.DetectContentType`.
func ParseHTML(r io.Reader, contentType, urlStr string) (*Microdata, error) {
	if contentType == "" {
		b := make([]byte, 512)
		n, err := r.Read(b)
		if err != nil && !errors.Is(err, io.EOF) {
			return nil, err
		}
		contentType = http.DetectContentType(b[:n])
		r = io.MultiReader(bytes.NewReader(b[:n]), r)
	}

	cr, err := charset.NewReader(r, contentType)
	if err != nil {
		return nil, err
	}

	tree, err := html.Parse(cr)
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
