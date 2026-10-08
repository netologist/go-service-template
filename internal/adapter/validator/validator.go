package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const tagParts = 2

func New(tagName string) *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get(tagName), ",", tagParts)[0]
		if name == "-" || name == "" {
			return f.Name
		}
		return name
	})
	return v
}
