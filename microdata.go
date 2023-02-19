// Copyright 2015 Lars Wiegman. All rights reserved. Use of this source code is
// governed by a BSD-style license that can be found in the LICENSE file.

/*

	Package flag implements command-line flag parsing.
	Package microdata implements a HTML microdata parser. It depends on the
	golang.org/x/net/html HTML5-compliant parser.

	Usage:

	Pass a reader, content-type and a base URL to the ParseHTML function.
		data, err := microdata.ParseHTML(reader, contentType, baseURL)
		items := data.Items

	Pass an URL to the ParseURL function.
		data, _ := microdata.ParseURL("http://example.com/blogposting")
		items := data.Items
*/

package microdata

import (
	"bytes"
	"io"
	"net/http"
	"net/url"
)

// ParseHTML parses the HTML document available in the given reader and returns
// the microdata. The given url is used to resolve the URLs in the
// attributes. The given contentType is used convert the content of r to UTF-8.
// When the given contentType is equal to "", the content type will be detected
// using `http.DetectContentType`.
func ParseHTML(r io.Reader, contentType string, u *url.URL) (*Microdata, error) {
	if contentType == "" {
		b := make([]byte, 512)
		_, err := r.Read(b)
		if err != nil {
			return nil, err
		}
		contentType = http.DetectContentType(b)
		r = io.MultiReader(bytes.NewReader(b), r)
	}

	p, err := newParser(r, contentType, u)
	if err != nil {
		return nil, err
	}
	return p.parse()
}

// ParseURL parses the HTML document available at the given URL and returns the
// microdata.
func ParseURL(urlStr string) (*Microdata, error) {
	var data *Microdata

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Get(urlStr)
	if err != nil {
		return data, err
	}

	contentType := resp.Header.Get("Content-Type")

	p, err := newParser(resp.Body, contentType, u)
	if err != nil {
		return nil, err
	}

	return p.parse()
}
