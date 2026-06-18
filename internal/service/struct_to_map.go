package service

import (
	"reflect"
	"slices"
)

func StructToMap(s any, execute []string) map[string]any {
	v := reflect.ValueOf(s)

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	res := make(map[string]any, v.NumField())
	for i := 0; i < v.NumField(); i++ {
		filed := t.Field(i)
		if !slices.Contains(execute, filed.Name) {
			res[filed.Name] = v.Field(i).Interface()
		}
	}

	return res
}
