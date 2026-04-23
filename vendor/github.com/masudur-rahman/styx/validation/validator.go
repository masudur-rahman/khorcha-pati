package validation

import (
	"reflect"

	"github.com/masudur-rahman/styx/dberr"
)

// Validate checks a struct's fields against validate struct tags.
// Returns nil if valid, or a *dberr.ValidationError with per-field errors.
func Validate(doc any) error {
	val := reflect.ValueOf(doc)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return nil
	}

	fieldErrors := make(map[string][]string)
	valType := val.Type()

	for i := 0; i < val.NumField(); i++ {
		field := valType.Field(i)
		tag := field.Tag.Get("validate")
		if tag == "" {
			continue
		}

		rules := ParseRules(tag)
		fieldValue := val.Field(i).Interface()

		for _, rule := range rules {
			if msg := ApplyRule(rule, fieldValue, field.Name); msg != "" {
				fieldErrors[field.Name] = append(fieldErrors[field.Name], msg)
			}
		}
	}

	if len(fieldErrors) > 0 {
		return dberr.NewValidationError(fieldErrors)
	}
	return nil
}
