package dberr

import "reflect"

func CheckEntityNameNonEmpty(entity string) error {
	if entity == "" {
		return ErrInvalidEntityName
	}
	return nil
}

func CheckIDNonEmpty(id any) error {
	if IsZeroValue(id) {
		return ErrInvalidID
	}
	return nil
}

func CheckIdOrFilterNonEmpty(id any, filter interface{}) error {
	if IsZeroValue(id) && filter == nil {
		return ErrInvalidID
	}
	return nil
}

// IsZeroValue checks if a value is its type's zero value.
func IsZeroValue(value any) bool {
	if value == nil {
		return true
	}
	typ := reflect.TypeOf(value)
	zero := reflect.Zero(typ).Interface()
	return reflect.DeepEqual(value, zero)
}
