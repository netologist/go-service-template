package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

const jsonTagParts = 2

func New() *validator.Validate {
	v := validator.New()
	v.RegisterTagNameFunc(func(f reflect.StructField) string {
		name := strings.SplitN(f.Tag.Get("json"), ",", jsonTagParts)[0]
		if name == "-" || name == "" {
			return f.Name
		}
		return name
	})
	return v
}

type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

func FormatValidationErrors(err error) []FieldError {
	var ve validator.ValidationErrors
	if !errors.As(err, &ve) {
		return []FieldError{{Message: "invalid request"}}
	}

	out := make([]FieldError, 0, len(ve))
	for _, fe := range ve {
		out = append(out, FieldError{
			Field:   fe.Field(),
			Message: messageFor(fe),
		})
	}
	return out
}

func messageFor(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("'%s' is required", fe.Field())
	case "email":
		return fmt.Sprintf("'%s' must be a valid email", fe.Field())
	case "min":
		return fmt.Sprintf("'%s' must be at least %s characters", fe.Field(), fe.Param())
	case "max":
		return fmt.Sprintf("'%s' must be at most %s characters", fe.Field(), fe.Param())
	case "uuid":
		return fmt.Sprintf("'%s' must be a valid UUID", fe.Field())
	default:
		return fmt.Sprintf("'%s' is invalid", fe.Field())
	}
}
