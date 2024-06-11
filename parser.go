package microdata

import (
	"bytes"
	"log"
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
func (p *parser) parse() (*Microdata, error) {
	var toplevelNodes []*html.Node
	var jsonNodes []*html.Node

	walkNodes(p.tree, func(n *html.Node) {
		if n.DataAtom == atom.Script && checkAttr("type", "application/ld+json", n) {
			jsonNodes = append(jsonNodes, n)
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
		p.readAttr(item, node)
		p.readItem(item, node, true)
	}

	for _, node := range jsonNodes {
		if node.FirstChild != nil {
			data := []byte(node.FirstChild.Data)

			var jsonMap interface{}
			err := fixjson.Unmarshal(data, &jsonMap)
			if err == nil {
				p.readJsonItem(nil, jsonMap)
			} else {
				log.Println("Error parsing json:", err)
			}
		}
	}

	return p.data, nil
}

func (p *parser) readJsonItem(item *Item, mi interface{}) {
	switch mi.(type) {
	case []interface{}: // assume this is array of items
		for _, i := range mi.([]interface{}) {
			p.readJsonItem(item, i)
		}
	case map[string]interface{}: // assume this is a root of an item
		m := mi.(map[string]interface{})

		if item == nil {
			item = NewItem()
			p.data.addItem(item)
		}

		if m["@type"] != nil {
			p.readType(item, m["@type"])
		}

		// sometimes they forget about @ char :/
		if m["type"] != nil {
			p.readType(item, m["type"])
		}

		for k, v := range m {
			p.readJsonProp(item, k, v)
		}
	}
}

func (p *parser) readType(item *Item, val interface{}) {
	switch vt := val.(type) {
	case []interface{}:
		for _, sv := range vt {
			p.readType(item, sv)
		}
	case string:
		item.addType(val.(string))
	}
}

// readJsonProp depending on value type, adds the value to the given item.
func (p *parser) readJsonProp(item *Item, key string, value interface{}) {
	if key == "@type" {
		return
	}

	switch vt := value.(type) {
	case []interface{}:
		for _, sv := range vt {
			p.readJsonProp(item, key, sv)
		}
	case map[string]interface{}:
		newItem := NewItem()
		item.addItem(key, newItem)
		p.readJsonItem(newItem, value)
	case nil:
	default:
		item.addProperty(key, value)
	}
}

// readItem traverses the given node tree, applying relevant attributes to the given item.
func (p *parser) readItem(item *Item, node *html.Node, isToplevel bool) {
	itemprops, hasProp := getAttr("itemprop", node)
	_, hasScope := getAttr("itemscope", node)

	switch {
	case hasScope && hasProp:
		subItem := NewItem()
		p.readAttr(subItem, node)
		for _, propName := range strings.Split(itemprops, " ") {
			if len(propName) > 0 {
				item.addItem(propName, subItem)
			}
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			p.readItem(subItem, c, false)
		}
		return
	case !hasScope && hasProp:
		if s := p.getValue(node); len(s) > 0 {
			for _, propName := range strings.Split(itemprops, " ") {
				if len(propName) > 0 {
					item.addProperty(propName, s)
				}
			}
		}
	case hasScope && !isToplevel:
		return
	}

	for c := node.FirstChild; c != nil; c = c.NextSibling {
		p.readItem(item, c, false)
	}
}

// readAttr applies relevant attributes from the given node to the given item.
func (p *parser) readAttr(item *Item, node *html.Node) {
	if s, ok := getAttr("itemtype", node); ok {
		for _, itemtype := range strings.Split(s, " ") {
			if len(itemtype) > 0 {
				item.addType(itemtype)
			}
		}

		if s, ok := getAttr("itemid", node); ok {
			if u, err := p.baseURL.Parse(s); err == nil {
				item.ID = u.String()
			}
		}
	}

	if s, ok := getAttr("itemref", node); ok {
		for _, itemref := range strings.Split(s, " ") {
			if len(itemref) > 0 {
				if n, ok := p.identifiedNodes[itemref]; ok {
					p.readItem(item, n, false)
				}
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
		value, ok := getAttr("data-src", node)
		if !ok {
			value, ok = getAttr("src", node)
		}

		if ok {
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
		// The "content" attribute can be found on other tags besides the meta tag.
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
func newParser(root *html.Node, baseURL *url.URL) (*parser, error) {
	return &parser{
		tree:            root,
		data:            &Microdata{},
		baseURL:         baseURL,
		identifiedNodes: make(map[string]*html.Node),
	}, nil
}
