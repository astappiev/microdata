# Microdata

Microdata is a package to extract [Microdata](https://www.w3.org/TR/microdata/) and [JSON-LD](https://www.w3.org/TR/json-ld/) from HTML documents.

__HTML Microdata__ is a markup specification often used in combination with the [schema collection](https://schema.org/docs/schemas.html) to make it easier for search engines to identify and understand content on web pages. One of the most common schemas is the rating you see when you google for something. Other schemas are persons, places, events, products, etc.

__JSON-LD__ is a lightweight Linked Data format. It is easy for humans to read and write. It is based on the already successful JSON format and provides a way to help JSON data interoperate at Web-scale.


## Go package use

Install the package:

```sh
go get -u github.com/astappiev/microdata
```

Use cases:
```go
// Pass a URL to the `ParseURL` function.
data, err := microdata.ParseURL("https://example.com/page")

// Pass a `io.Reader`, content-type and a base URL to the `ParseHTML` function.
data, err := microdata.ParseHTML(reader, contentType, baseURL)

// Pass a `html.Node`, content-type and a base URL to the `ParseNode` function.
data, err := microdata.ParseNode(reader, contentType, baseURL)
```

An example program:
```go
package main

import (
    "encoding/json"
    "fmt"

    "github.com/astappiev/microdata"
)

func main() {
    data, _ := microdata.ParseURL("https://www.allrecipes.com/recipe/84450/ukrainian-red-borscht-soup/")
    
    // iterate over metadata items:
    items := data.Items
	for _, item := range items {
		fmt.Println(item.Types)
		for key, prop := range item.Properties {
			fmt.Printf("%s: %v\n", key, prop)
		}
	}

    // print json schema
    json, _ := json.MarshalIndent(data, "", "  ")
    fmt.Println(string(json))
}
```


## Command line use

Install the command line tool:

```sh
go get -u github.com/astappiev/microdata/cmd/microdata
```

Parse an URL:

```sh
microdata https://www.gog.com/game/...
{
  "items": [
    {
      "type": [
        "http://schema.org/Product"
      ],
      "properties": {
        "additionalProperty": [
          {
            "type": [
              "http://schema.org/PropertyValue"
            ],
{
...
```

Parse HTML from the stdin:

```
$ cat saved.html | microdata
```

Format the output with a Go template to return the "price" property:

```sh
microdata -format '{{with index .Items 0}}{{with index .Properties "offers" 0}}{{with index .Properties "price" 0 }}{{ . }}{{end}}{{end}}{{end}}' https://www.gog.com/game/...
8.99
```
