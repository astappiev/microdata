// Package microdata is a package to extract Microdata and JSON-LD from HTML documents.
//
// HTML Microdata is a markup specification often used in combination with the schema collection
// to make it easier for search engines to identify and understand content on web pages.
//
// JSON-LD is a lightweight Linked Data format. It is easy for humans to read and write.
// It is based on the already successful JSON format and provides a way to help JSON data
// interoperate at Web-scale.
//
// Use cases:
//
//	// Pass a URL to the ParseURL function.
//	data, err := microdata.ParseURL("https://example.com/page")
//
//	// Pass a context and a URL to the ParseURLWithContext function.
//	data, err := microdata.ParseURLWithContext(ctx, "https://example.com/page")
//
//	// Pass a io.Reader, content-type and a base URL to the ParseHTML function.
//	data, err := microdata.ParseHTML(reader, contentType, baseURL)
//
//	// Pass a html.Node and a base URL to the ParseNode function.
//	data, err := microdata.ParseNode(rootNode, baseURL)
//
// Note: The parser tolerates JSON-LD "type" attributes without the leading "@" character,
// and it intentionally drops empty properties instead of recording empty strings.
package microdata
