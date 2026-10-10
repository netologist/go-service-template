package config

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

// NewValidation creates a new validator instance with tag name registration.
func NewValidation(tagName string) *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name, _, _ := strings.Cut(f.Tag.Get(tagName), ",")
		if name == "-" || name == "" {
			return f.Name
		}
		return name
	})
	return v
}
