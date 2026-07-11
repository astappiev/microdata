package microdata

import (
	"bytes"
	"net/url"
	"strings"

	"github.com/astappiev/fixjson"
	"golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

type parser struct {
	tree            *html.Node
	data            *Microdata
	baseURL         *url.URL
	identifiedNodes map[string]*html.Node
}

// parse returns the microdata from the parser's node tree.
func (p *parser) parse() *Microdata {
	var toplevelNodes []*html.Node
	var jsonNodes []*html.Node

	walkNodes(p.tree, func(n *html.Node) {
		if n.DataAtom == atom.Script {
			if t, ok := getAttr("type", n); ok {
				if idx := strings.IndexByte(t, ';'); idx >= 0 {
					t = t[:idx]
				}
				if strings.EqualFold(strings.TrimSpace(t), "application/ld+json") {
					jsonNodes = append(jsonNodes, n)
				}
			}
		}

		if _, ok := getAttr("itemscope", n); ok {
			if _, ok := getAttr("itemprop", n); !ok {
				toplevelNodes = append(toplevelNodes, n)
			}
		}

		if id, ok := getAttr("id", n); ok {
			p.identifiedNodes[id] = n
		}
	})

	for _, node := range toplevelNodes {
		item := NewItem()
		p.data.addItem(item)
		visited := make(map[*html.Node]bool)
		p.readAttr(item, node, visited)
		p.readItem(item, node, true, visited)
	}

	for _, node := range jsonNodes {
		if node.FirstChild != nil {
			data := []byte(node.FirstChild.Data)

			var jsonMap any
			err := fixjson.Unmarshal(data, &jsonMap)
			if err == nil {
				p.readJSONItem(nil, jsonMap)
			}
		}
	}

	return p.data
}

func (p *parser) readJSONItem(item *Item, value any) {
	switch v := value.(type) {
	case []any: // assume this is array of items
		for _, i := range v {
			p.readJSONItem(item, i)
		}
	case map[string]any: // assume this is a root of an item
		if item == nil {
			item = NewItem()
			p.data.addItem(item)
		}

		if v["@type"] != nil {
			p.readType(item, v["@type"])
		}

		// sometimes they forget about @ char :/
		if v["type"] != nil {
			p.readType(item, v["type"])
		}

		for k, val := range v {
			p.readJSONProp(item, k, val)
		}
	}
}

func (p *parser) readType(item *Item, val any) {
	switch vt := val.(type) {
	case []any:
		for _, sv := range vt {
			p.readType(item, sv)
		}
	case string:
		item.addType(vt)
	}
}

// readJSONProp depending on value type, adds the value to the given item.
func (p *parser) readJSONProp(item *Item, key string, value any) {
	if key == "@type" {
		return
	}

	if key == "type" {
		switch value.(type) {
		case string, []any:
			return
		}
	}

	switch vt := value.(type) {
	case []any:
		for _, sv := range vt {
			p.readJSONProp(item, key, sv)
		}
	case map[string]any:
		newItem := NewItem()
		item.addProperty(key, newItem)
		p.readJSONItem(newItem, value)
	case nil:
	default:
		item.addProperty(key, value)
	}
}

// readItem traverses the given node tree, applying relevant attributes to the given item.
func (p *parser) readItem(item *Item, node *html.Node, isToplevel bool, visited map[*html.Node]bool) {
	if visited[node] {
		return
	}
	visited[node] = true
	defer delete(visited, node)

	itemprops, hasProp := getAttr("itemprop", node)
	_, hasScope := getAttr("itemscope", node)

	switch {
	case hasScope && hasProp:
		subItem := NewItem()
		p.readAttr(subItem, node, visited)
		for propName := range strings.FieldsSeq(itemprops) {
			item.addProperty(propName, subItem)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			p.readItem(subItem, c, false, visited)
		}
		return
	case !hasScope && hasProp:
		if s := p.getValue(node); len(s) > 0 {
			for propName := range strings.FieldsSeq(itemprops) {
				item.addProperty(propName, s)
			}
		}
	case hasScope && !isToplevel:
		return
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.readItem(item, c, false, visited)
	}
}

// readAttr applies relevant attributes from the given node to the given item.
func (p *parser) readAttr(item *Item, node *html.Node, visited map[*html.Node]bool) {
	if s, ok := getAttr("itemtype", node); ok {
		for itemtype := range strings.FieldsSeq(s) {
			item.addType(itemtype)
		}

		if s, ok := getAttr("itemid", node); ok {
			if u, err := p.baseURL.Parse(s); err == nil {
				item.ID = u.String()
			}
		}
	}

	if s, ok := getAttr("itemref", node); ok {
		for itemref := range strings.FieldsSeq(s) {
			if n, ok := p.identifiedNodes[itemref]; ok {
				p.readItem(item, n, false, visited)
			}
		}
	}
}

// getValue returns the value of the property, value pair in the given node.
func (p *parser) getValue(node *html.Node) string {
	var propValue string

	switch node.DataAtom {
	case atom.Meta:
		if value, ok := getAttr("content", node); ok {
			propValue = value
		}
	case atom.Audio, atom.Embed, atom.Iframe, atom.Source, atom.Track, atom.Video:
		if value, ok := getAttr("src", node); ok {
			if u, err := p.baseURL.Parse(value); err == nil {
				propValue = u.String()
			}
		}
	case atom.Img:
		// Prefer data-src over src for lazy-loading support
		value, ok := getAttr("data-src", node)
		if !ok {
			value, ok = getAttr("src", node)
		}

		if ok {
			if u, err := p.baseURL.Parse(value); err == nil {
				propValue = u.String()
			}
		}
	case atom.Object:
		if value, ok := getAttr("data", node); ok {
			if u, err := p.baseURL.Parse(value); err == nil {
				propValue = u.String()
			}
		}
	case atom.A, atom.Area, atom.Link:
		if value, ok := getAttr("href", node); ok {
			if u, err := p.baseURL.Parse(value); err == nil {
				propValue = u.String()
			}
		}
	case atom.Data, atom.Meter:
		if value, ok := getAttr("value", node); ok {
			propValue = value
		}
	case atom.Time:
		if value, ok := getAttr("datetime", node); ok {
			propValue = value
		}
	default:
		// The "content" attribute can be found on other tags besides the meta tag,
		// and is used before falling back to text content (RDFa-style).
		if value, ok := getAttr("content", node); ok {
			propValue = value
			break
		}

		var buf bytes.Buffer
		walkNodes(node, func(n *html.Node) {
			if n.Type == html.TextNode {
				buf.WriteString(n.Data)
			}
		})
		propValue = buf.String()
	}

	return propValue
}

// newParser returns a parser that converts the contents of the given node tree to microdata.
func newParser(root *html.Node, baseURL *url.URL) *parser {
	return &parser{
		tree:            root,
		data:            &Microdata{},
		baseURL:         baseURL,
		identifiedNodes: make(map[string]*html.Node),
	}
}
