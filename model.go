package microdata

import (
	"fmt"
	"slices"
)

// Microdata represents the extracted microdata from a HTML document.
type Microdata struct {
	Items []*Item `json:"items"`
}

// addItem adds the item to the item list.
func (m *Microdata) addItem(item *Item) {
	m.Items = append(m.Items, item)
}

// GetFirstOfSchemaType returns the first item of the given type with a possible https://schema.org/ context.
func (m *Microdata) GetFirstOfSchemaType(itemType string) *Item {
	return m.GetFirstOfType(itemType, "http://schema.org/"+itemType, "https://schema.org/"+itemType)
}

// GetFirstOfType returns the first item of the given type.
func (m *Microdata) GetFirstOfType(itemType ...string) *Item {
	for _, item := range m.Items {
		for _, t := range item.Types {
			if slices.Contains(itemType, t) {
				return item
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

// ValueList represents a list of values for a property.
type ValueList []any

// PropertyMap represents a map of property names to their corresponding value lists.
type PropertyMap map[string]ValueList

// Item represents a single microdata item.
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
func (i *Item) addProperty(key string, value any) {
	i.Properties[key] = append(i.Properties[key], value)
}

// IsOfSchemaType returns whether the item is of the given type within the schema.org context.
func (i *Item) IsOfSchemaType(itemType string) bool {
	return i.IsOfType(itemType, "http://schema.org/"+itemType, "https://schema.org/"+itemType)
}

// IsOfType returns whether the item is of the given type.
func (i *Item) IsOfType(itemType ...string) bool {
	for _, t := range i.Types {
		if slices.Contains(itemType, t) {
			return true
		}
	}
	return false
}

// GetProperty returns the first value of the first key that has at least one value.
func (i *Item) GetProperty(keys ...string) (val any, ok bool) {
	for _, key := range keys {
		if arr, ok := i.GetProperties(key); ok {
			return arr[0], true
		}
	}
	return nil, false
}

// GetProperties returns the values of the first key that has at least one value.
func (i *Item) GetProperties(keys ...string) (arr []any, ok bool) {
	for _, key := range keys {
		if props := i.Properties[key]; len(props) > 0 {
			return props, true
		}
	}
	return nil, false
}

// GetNestedItem returns the first item from the properties of the first key that has at least one value.
func (i *Item) GetNestedItem(keys ...string) (val *Item, ok bool) {
	if data, ok := i.GetNested(keys...); ok {
		return data.Items[0], true
	}
	return nil, false
}

// GetNested returns the extracted microdata for the properties of the first key that has at least one value.
func (i *Item) GetNested(keys ...string) (data Microdata, ok bool) {
	for _, key := range keys {
		var arr []*Item
		for _, v := range i.Properties[key] {
			if item, ok := v.(*Item); ok {
				arr = append(arr, item)
			}
		}
		if len(arr) > 0 {
			return Microdata{Items: arr}, true
		}
	}
	return Microdata{}, false
}

// CountPaths recursively counts the occurrences of item paths, storing the results in the provided paths map.
func (i *Item) CountPaths(prefix string, paths map[string]int) {
	for key, val := range i.Properties {
		for _, vv := range val {
			paths[fmt.Sprintf("%s[%T]", prefix+key, vv)]++

			if item, ok := vv.(*Item); ok {
				item.CountPaths(prefix+key+".", paths)
			}
		}
	}
}

func NewItem() *Item {
	return &Item{
		Types:      make([]string, 0),
		Properties: make(PropertyMap),
	}
}
