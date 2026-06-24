package service

import (
	"reflect"
	"slices"
	"strings"
)

func StructToMap(s any, execute []string) map[string]any {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}

	t := v.Type()

	res := make(map[string]any, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		key := structFieldKey(field)
		if key == "-" {
			continue
		}
		if slices.Contains(execute, key) || slices.Contains(execute, field.Name) {
			continue
		}

		value := v.Field(i)
		if !value.IsZero() {
			res[key] = value.Interface()
		}
	}

	return res
}

func structFieldKey(field reflect.StructField) string {
	for _, tagName := range []string{"db", "json"} {
		tagValue, ok := field.Tag.Lookup(tagName)
		if !ok {
			continue
		}

		name, _, _ := strings.Cut(tagValue, ",")
		if name != "" {
			return name
		}
	}

	return field.Name
}
