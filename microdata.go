package microdata

import (
	"bytes"
	"context"
	"errors"
	"fmt"
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
// It uses http.DefaultClient to fetch the document.
func ParseURLWithContext(ctx context.Context, urlStr string) (*Microdata, error) {
	return ParseURLWithClient(ctx, http.DefaultClient, urlStr)
}

// ParseURLWithClient parses the HTML document available at the given URL using the provided context and client.
func ParseURLWithClient(ctx context.Context, client *http.Client, urlStr string) (*Microdata, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
	if err != nil {
		return nil, err
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("microdata: fetch %s: unexpected status %s", urlStr, resp.Status)
	}

	contentType := resp.Header.Get("Content-Type")
	return ParseHTML(resp.Body, contentType, resp.Request.URL.String())
}

// ParseHTML parses the HTML document available in the given reader and returns the microdata. The given url is
// used to resolve the URLs in the attributes. The given contentType is used to convert the content of r to UTF-8.
// When the given contentType is equal to "", the content type will be detected using `http.DetectContentType`.
func ParseHTML(r io.Reader, contentType, urlStr string) (*Microdata, error) {
	if contentType == "" {
		b := make([]byte, 512)
		n, err := io.ReadFull(r, b)
		if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
			return nil, err
		}
		contentType = http.DetectContentType(b[:n])
		r = io.MultiReader(bytes.NewReader(b[:n]), r)
	}

	cr, err := charset.NewReader(r, contentType)
	if err != nil {
		return nil, fmt.Errorf("microdata: read input: %w", err)
	}

	tree, err := html.Parse(cr)
	if err != nil {
		return nil, fmt.Errorf("microdata: parse html: %w", err)
	}

	return ParseNode(tree, urlStr)
}

// ParseNode parses the root Node and returns the microdata.
func ParseNode(root *html.Node, urlStr string) (*Microdata, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	return newParser(root, u).parse(), nil
}
