package microdata

type Microdata struct {
	Items []*Item `json:"items"`
}

// addItem adds the item to the items list.
func (m *Microdata) addItem(item *Item) {
	m.Items = append(m.Items, item)
}

type ValueList []interface{}

type PropertyMap map[string]ValueList

type Item struct {
	Types      []string    `json:"type"`
	Properties PropertyMap `json:"properties"`
	ID         string      `json:"id,omitempty"`
}

// addString adds the property, value pair to the properties map. It appends to any
// existing property.
func (i *Item) addString(property, value string) {
	i.Properties[property] = append(i.Properties[property], value)
}

// addItem adds the property, value pair to the properties map. It appends to any
// existing property.
func (i *Item) addItem(property string, value *Item) {
	i.Properties[property] = append(i.Properties[property], value)
}

// addType adds the value to the types list.
func (i *Item) addType(value string) {
	i.Types = append(i.Types, value)
}

// NewItem returns a new Item.
func NewItem() *Item {
	return &Item{
		Types:      make([]string, 0),
		Properties: make(PropertyMap, 0),
	}
}
