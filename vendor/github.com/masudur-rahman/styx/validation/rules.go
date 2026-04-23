package validation

import (
	"fmt"
	"net/mail"
	"net/url"
	"strconv"
	"strings"
	"unicode"
)

// Rule represents a single validation rule parsed from a struct tag.
type Rule struct {
	Name  string
	Param string
}

// ParseRules parses a validate tag value like "required,min:3,max:100,email".
func ParseRules(tag string) []Rule {
	if tag == "" {
		return nil
	}
	var rules []Rule
	for _, part := range strings.Split(tag, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		name, param, _ := strings.Cut(part, ":")
		rules = append(rules, Rule{Name: name, Param: param})
	}
	return rules
}

// ApplyRule validates a field value against a single rule.
// Returns an error message if validation fails, empty string otherwise.
func ApplyRule(rule Rule, value any, fieldName string) string {
	switch rule.Name {
	case "required":
		return checkRequired(value, fieldName)
	case "min":
		return checkMin(value, rule.Param, fieldName)
	case "max":
		return checkMax(value, rule.Param, fieldName)
	case "email":
		return checkEmail(value, fieldName)
	case "url":
		return checkURL(value, fieldName)
	case "numeric":
		return checkNumeric(value, fieldName)
	case "alpha":
		return checkAlpha(value, fieldName)
	default:
		return ""
	}
}

func checkRequired(value any, field string) string {
	if value == nil {
		return fmt.Sprintf("%s is required", field)
	}
	if s, ok := value.(string); ok && strings.TrimSpace(s) == "" {
		return fmt.Sprintf("%s is required", field)
	}
	return ""
}

func checkMin(value any, param string, field string) string {
	n, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		if int64(len(v)) < n {
			return fmt.Sprintf("%s must be at least %d characters", field, n)
		}
	case int:
		if int64(v) < n {
			return fmt.Sprintf("%s must be at least %d", field, n)
		}
	case int64:
		if v < n {
			return fmt.Sprintf("%s must be at least %d", field, n)
		}
	}
	return ""
}

func checkMax(value any, param string, field string) string {
	n, err := strconv.ParseInt(param, 10, 64)
	if err != nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		if int64(len(v)) > n {
			return fmt.Sprintf("%s must be at most %d characters", field, n)
		}
	case int:
		if int64(v) > n {
			return fmt.Sprintf("%s must be at most %d", field, n)
		}
	case int64:
		if v > n {
			return fmt.Sprintf("%s must be at most %d", field, n)
		}
	}
	return ""
}

func checkEmail(value any, field string) string {
	s, ok := value.(string)
	if !ok || s == "" {
		return ""
	}
	if _, err := mail.ParseAddress(s); err != nil {
		return fmt.Sprintf("%s must be a valid email", field)
	}
	return ""
}

func checkURL(value any, field string) string {
	s, ok := value.(string)
	if !ok || s == "" {
		return ""
	}
	if u, err := url.Parse(s); err != nil || u.Scheme == "" || u.Host == "" {
		return fmt.Sprintf("%s must be a valid URL", field)
	}
	return ""
}

func checkNumeric(value any, field string) string {
	s, ok := value.(string)
	if !ok || s == "" {
		return ""
	}
	for _, r := range s {
		if !unicode.IsDigit(r) && r != '.' && r != '-' {
			return fmt.Sprintf("%s must be numeric", field)
		}
	}
	return ""
}

func checkAlpha(value any, field string) string {
	s, ok := value.(string)
	if !ok || s == "" {
		return ""
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return fmt.Sprintf("%s must contain only letters", field)
		}
	}
	return ""
}
