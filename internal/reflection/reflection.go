package reflection

import (
	"reflect"
	"strings"
)

// ToAttributes accepts a struct (probably your service request) and returns a list
// of the key/value pairs for the attribute name/values. This is recursive, so nested
// structs will be included as the 'Value' of the necessary attributes.
func ToAttributes(item interface{}) StructAttributes {
	var attrs StructAttributes
	if item == nil {
		return attrs
	}

	valueType := reflect.TypeOf(item)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(item))

	for i := 0; i < valueType.NumField(); i++ {
		name := BindingName(valueType.Field(i))
		field := reflectValue.Field(i).Interface()
		if valueType.Field(i).Type.Kind() == reflect.Struct {
			attrs = append(attrs, &StructAttribute{Name: name, Value: ToAttributes(field)})
		} else {
			attrs = append(attrs, &StructAttribute{Name: name, Value: field})
		}
	}
	return attrs
}

// StructAttributes maintains a list of attribute names/values for some source struct.
type StructAttributes []*StructAttribute

// Find looks up the struct attribute with the given name. This search is CASE-INSENSITIVE
// in order to match how the standard library handles JSON attribute.
func (attrs StructAttributes) Find(name string) *StructAttribute {
	for _, attr := range attrs {
		if attr.Matches(name) {
			return attr
		}
	}
	return nil
}

// Remove performs a case-insensitive search for the attribute w/ that name and removes it from
// the list. The new slice without the matching attribute is returned.
func (attrs StructAttributes) Remove(name string) StructAttributes {
	for i, attr := range attrs {
		if attr.Matches(name) {
			return append(attrs[:i], attrs[i+1:]...)
		}
	}
	return attrs
}

// StructAttribute represents a single key/value pair for a field on a struct.
type StructAttribute struct {
	// Name is the binding name of the struct value. If there was a JSON tag on the
	// struct field, it should be that value. Otherwise it's just the struct field's name.
	Name string
	// Value is the runtime value of this field on the struct you ran through "ToAttributes()"
	Value interface{}
}

// Matches determines if there is a case-insensitive match between this name and the field.
func (attr StructAttribute) Matches(name string) bool {
	return strings.EqualFold(name, attr.Name)
}

var noField = reflect.StructField{}

// FindField looks up the struct field attribute for the given field on the given struct.
func FindField(structType reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(name, BindingName(field)) {
			return field, true
		}
	}
	return noField, false
}

// BindingName just returns the name of the field/attribute on the struct unless it has a `json` tag
// defined. If so, it will use the remapped name for this field instead.
//
//     type Foo struct {
//         A string
//         B string `json:"hello"
//     }
//
// The binding name for the first attribute is "A", but the binding name for the other is "hello".
func BindingName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "" || jsonTag == "-" {
		return field.Name
	}

	// Parse the `json` tag to determine how the user has re-mapped the field.
	switch comma := strings.IndexRune(jsonTag, ','); comma {
	case -1:
		// e.g. `json:"firstName"`
		return jsonTag
	case 0:
		// e.g. `json:",omitempty"` (not remapped so use fields actual name)
		return field.Name
	default:
		// e.g. `json:"firstName,omitempty" (just use the remapped name)
		return jsonTag[0:comma]
	}
}

// Assign simply performs a reflective replace of the value, making sure to try to properly handle pointers.
func Assign(value interface{}, out interface{}) {
	// Depending on whether you wrote "SomeStruct{}" or "&SomeStruct{}" (a pointer) to the
	// scope, we want to make sure that we're de-referencing the scope value properly.
	reflectValue := reflect.ValueOf(value)
	if reflectValue.Type().Kind() == reflect.Ptr {
		reflect.ValueOf(out).Elem().Set(reflectValue.Elem())
	} else {
		reflect.ValueOf(out).Elem().Set(reflectValue)
	}
}
