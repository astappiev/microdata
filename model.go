package microdata

import (
	"fmt"
)

type Microdata struct {
	Items []*Item `json:"items"`
}

// addItem adds the item to the items list.
func (m *Microdata) addItem(item *Item) {
	m.Items = append(m.Items, item)
}

// GetFirstOfSchemaType returns the first item of the given type with possible https://schema.org/ context.
func (m *Microdata) GetFirstOfSchemaType(itemType string) *Item {
	return m.GetFirstOfType(itemType, "http://schema.org/"+itemType, "https://schema.org/"+itemType)
}

// GetFirstOfType returns the first item of the given type.
func (m *Microdata) GetFirstOfType(itemType ...string) *Item {
	for _, item := range m.Items {
		for _, t1 := range item.Types {
			for _, t2 := range itemType {
				if t1 == t2 {
					return item
				}
			}
		}

		if graph, ok := item.GetNested("@graph"); ok {
			if item := graph.GetFirstOfType(itemType...); item != nil {
				return item
			}
		}
	}

	return nil
}

type ValueList []interface{}

type PropertyMap map[string]ValueList

type Item struct {
	Types      []string    `json:"type"`
	Properties PropertyMap `json:"properties"`
	ID         string      `json:"id,omitempty"`
}

// addType adds the value to the types list.
func (i *Item) addType(value string) {
	i.Types = append(i.Types, value)
}

// addProperty adds the property, value pair to the properties map. It appends to any existing property.
func (i *Item) addProperty(key string, value interface{}) {
	i.Properties[key] = append(i.Properties[key], value)
}

// addItem adds the property, value pair to the properties map. It appends to any existing property.
func (i *Item) addItem(key string, value *Item) {
	i.Properties[key] = append(i.Properties[key], value)
}

func (i *Item) IsOfSchemaType(itemType string) bool {
	return i.IsOfType(itemType, "http://schema.org/"+itemType, "https://schema.org/"+itemType)
}

func (i *Item) IsOfType(itemType ...string) bool {
	for _, t1 := range i.Types {
		for _, t2 := range itemType {
			if t1 == t2 {
				return true
			}
		}
	}
	return false
}

func (i *Item) GetProperty(keys ...string) (val interface{}, ok bool) {
	for _, key := range keys {
		if arr, ok := i.GetProperties(key); ok {
			return arr[0], true
		}
	}
	return
}

func (i *Item) GetProperties(keys ...string) (arr []interface{}, ok bool) {
	for _, key := range keys {
		for _, v := range i.Properties[key] {
			arr = append(arr, v)
		}

		if len(arr) > 0 {
			return arr, true
		}
	}

	return arr, false
}

func (i *Item) GetNestedItem(keys ...string) (val *Item, ok bool) {
	if data, ok := i.GetNested(keys...); ok {
		return data.Items[0], true
	}
	return
}

func (i *Item) GetNested(keys ...string) (data Microdata, ok bool) {
	for _, key := range keys {
		var arr []*Item
		for _, v := range i.Properties[key] {
			switch v.(type) {
			case *Item:
				arr = append(arr, v.(*Item))
			}
		}
		return Microdata{Items: arr}, len(arr) > 0
	}
	return
}

func (i *Item) CountPaths(prefix string, paths *map[string]int) {
	for key, val := range i.Properties {
		(*paths)[fmt.Sprintf("%s[%T]", prefix+key, val[0])]++

		for _, vv := range val {
			switch vv.(type) {
			case *Item:
				vv.(*Item).CountPaths(prefix+key+".", paths)
			}
		}
	}
}

// NewItem returns a new Item.
func NewItem() *Item {
	return &Item{
		Types:      make([]string, 0),
		Properties: make(PropertyMap, 0),
	}
}
