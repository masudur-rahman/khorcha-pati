package dberr

import "github.com/masudur-rahman/styx/sql/postgres/lib"

func CheckEntityNameNonEmpty(entity string) error {
	if entity == "" {
		return NewRequirementMissing("entity name must be set")
	}
	return nil
}

func CheckIDNonEmpty(id any) error {
	if lib.IsZeroValue(id) {
		return NewRequirementMissing("must provide document id")
	}
	return nil
}

func CheckIdOrFilterNonEmpty(id any, filter interface{}) error {
	if lib.IsZeroValue(id) && filter == nil {
		return NewRequirementMissing("must provide id and/or filter")
	}
	return nil
}
