package hw09structvalidator

import (
	"errors"
	"reflect"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

var (
	ErrInvalidStringLength = errors.New("invalid string length")
	ErrValueMin            = errors.New("value too small")
	ErrValueMax            = errors.New("value too large")
	ErrInvalidItem         = errors.New("invalid item")
	ErrInvalidReqExp       = errors.New("invalid regexp")
)

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	result := ""
	for i, err := range v {
		if i > 0 {
			result += "; "
		}
		result += err.Field + ": " + err.Err.Error()
	}
	return result
}

func Validate(v interface{}) error {
	errors := ValidationErrors{}

	// if v is not a struct, return nil
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Struct {
		return nil
	}

	// if v is a struct, validate each field
	if rv.Kind() == reflect.Struct {
		types := rv.Type()

		for i := 0; i < rv.NumField(); i++ {
			field := types.Field(i)

			// validate field
			err := ValidateField(field, rv.Field(i))
			if err != nil {
				errors = append(errors, ValidationError{Field: field.Name, Err: err})
			}
		}
	}

	// if no errors, return nil
	if len(errors) == 0 {
		return nil
	}
	return errors
}

// validate field.
func ValidateField(field reflect.StructField, value reflect.Value) error {
	validation := field.Tag.Get("validate")
	if validation == "" {
		return nil
	}

	switch value.Type().Kind() {
	case reflect.String:
		return validateString(value.String(), validation)
	case reflect.Int:
		return validateInt(int(value.Int()), validation)
	case reflect.Slice:
		for i := range value.Seq() {
			err := ValidateField(field, value.Index(int(i.Int())))
			if err != nil {
				return err
			}
		}
	case reflect.Invalid, reflect.Bool, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64, reflect.Complex64, reflect.Complex128, reflect.Array,
		reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Struct,
		reflect.UnsafePointer:
		return nil
	default:
		return nil
	}

	return nil
}

// get validation rules.
func getValidationRules(validation string) map[string]interface{} {
	rules := map[string]interface{}{}

	// split validation into rules
	for _, rule := range strings.Split(validation, "|") {
		kv := strings.Split(rule, ":")
		switch kv[0] {
		case "len", "min", "max":
			v, err := strconv.Atoi(kv[1])
			if err == nil {
				rules[kv[0]] = v
			}
		case "in":
			rules[kv[0]] = strings.Split(kv[1], ",")
		case "regexp":
			rules[kv[0]] = kv[1]
		}
	}

	return rules
}

// validateString validates a string against a set of rules.
func validateString(value string, validation string) error {
	rules := getValidationRules(validation)

	for rule, constraint := range rules {
		switch rule {
		// validate string length
		case "len":
			if len(value) != constraint.(int) {
				return ErrInvalidStringLength
			}
		// validate string in list
		case "in":
			if !slices.Contains(constraint.([]string), value) {
				return ErrInvalidItem
			}
		// validate regexp
		case "regexp":
			matched, err := regexp.MatchString(constraint.(string), value)
			if err != nil {
				return err
			}
			if !matched {
				return ErrInvalidReqExp
			}
		}
	}
	return nil
}

// validateInt validates an int against a set of rules.
func validateInt(value int, validation string) error {
	rules := getValidationRules(validation)

	for rule, constraint := range rules {
		switch rule {
		case "min":
			if value < constraint.(int) {
				return ErrValueMin
			}
		case "max":
			if value > constraint.(int) {
				return ErrValueMax
			}
		case "in":
			if !slices.Contains(constraint.([]string), strconv.Itoa(value)) {
				return ErrInvalidItem
			}
		}
	}

	return nil
}
