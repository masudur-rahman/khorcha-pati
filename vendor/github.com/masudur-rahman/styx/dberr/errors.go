package dberr

import "errors"

type RequirementMissing struct {
	Message string
}

func (e RequirementMissing) Error() string {
	return e.Message
}

func NewRequirementMissing(msg string) RequirementMissing {
	return RequirementMissing{Message: msg}
}

var DataNotFound = errors.New("data not found")
