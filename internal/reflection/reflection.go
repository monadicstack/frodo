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

var noField = reflect.StructField{}

func FindField(structType reflect.Type, name string) (reflect.StructField, bool) {
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		if strings.EqualFold(name, FieldName(field)) {
			return field, true
		}
	}
	return noField, false
}

func FieldName(field reflect.StructField) string {
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
