package reflection

import (
	"reflect"
	"strings"
)

func StructToMap(item interface{}) map[string]interface{} {
	res := map[string]interface{}{}
	if item == nil {
		return res
	}

	valueType := reflect.TypeOf(item)
	if valueType.Kind() == reflect.Ptr {
		valueType = valueType.Elem()
	}

	reflectValue := reflect.Indirect(reflect.ValueOf(item))

	for i := 0; i < valueType.NumField(); i++ {
		name := FieldName(valueType.Field(i))
		field := reflectValue.Field(i).Interface()
		if valueType.Field(i).Type.Kind() == reflect.Struct {
			res[name] = StructToMap(field)
		} else {
			res[name] = field
		}
	}
	return res
}

func FieldName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "" || tag == "-" {
		return field.Name
	}

	switch comma := strings.IndexRune(tag, ','); {
	case comma == 0:
		// Ill-formed JSON tag
		return field.Name

	case comma >= 0:
		// It's something like "json:name,omitempty", so strip off just the remapped name, not the other options.
		return tag[0:comma]

	default:
		// The whole json tag is the remapped field name (e.g. "json:name")
		return tag
	}
}

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
