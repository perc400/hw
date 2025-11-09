package hw09structvalidator

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var (
	ErrArgumentNotStructure     = errors.New("not a structure")
	ErrFieldUnsupportedKind     = errors.New("unsupported field's kind")
	ErrFieldTagEmptyValue       = errors.New("tag key has empty value")
	ErrFieldStringInvalidLength = errors.New("invalid string length")
	ErrFieldStringNotMatchRegex = errors.New("string doesn't match regex")
	ErrFieldStringNotInSet      = errors.New("string not in set")
	ErrFieldIntNotInSet         = errors.New("int not in set")
	ErrFieldIntBiggerThanMax    = errors.New("int bigger than max value")
	ErrFieldIntLessThanMin      = errors.New("int less than min value")
)

type stringValidatorFunc func(value string, param string) error

var stringValidators = map[string]stringValidatorFunc{
	"len": func(value, param string) error {
		length, err := strconv.Atoi(param)
		if err != nil {
			return err
		}
		return validateLen(value, length)
	},
	"regexp": validateRegex,
	"in": func(value, param string) error {
		set := strings.Split(param, ",")
		return validateInString(value, set)
	},
}

type intValidatorFunc func(value int64, param string) error

var intValidators = map[string]intValidatorFunc{
	"max": func(value int64, param string) error {
		maxValue, err := strconv.Atoi(param)
		if err != nil {
			return err
		}
		return validateMax(value, maxValue)
	},
	"min": func(value int64, param string) error {
		minValue, err := strconv.Atoi(param)
		if err != nil {
			return err
		}
		return validateMin(value, minValue)
	},
	"in": func(value int64, param string) error {
		set := strings.Split(param, ",")
		return validateInInt(value, set)
	},
}

func validateMax(num int64, maxValue int) error {
	if num > int64(maxValue) {
		return ErrFieldIntBiggerThanMax
	}
	return nil
}

func validateMin(num int64, minValue int) error {
	if num < int64(minValue) {
		return ErrFieldIntLessThanMin
	}
	return nil
}

func validateInInt(num int64, set []string) error {
	found := false
	for _, part := range set {
		val, err := strconv.ParseInt(part, 10, 64)
		if err != nil {
			return err
		}
		if num == val {
			found = true
			break
		}
	}

	if !found {
		return ErrFieldIntNotInSet
	}
	return nil
}

func validateLen(str string, n int) error {
	if len(str) != n {
		return ErrFieldStringInvalidLength
	}
	return nil
}

func validateRegex(str, pattern string) error {
	matched, err := regexp.MatchString(pattern, str)
	if err != nil {
		return err
	}
	if !matched {
		return ErrFieldStringNotMatchRegex
	}
	return nil
}

func validateInString(str string, set []string) error {
	for _, v := range set {
		if str == v {
			return nil
		}
	}
	return ErrFieldStringNotInSet
}

func parseTag(tag string) (string, string, error) {
	tagParts := strings.SplitN(tag, ":", 2)
	if len(tagParts) < 2 {
		return "", "", ErrFieldTagEmptyValue
	}
	return tagParts[0], tagParts[1], nil
}

func checkStringField(fieldName string, fieldValue reflect.Value, tags []string) ValidationErrors {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   ErrFieldTagEmptyValue,
			})
			continue
		}

		fn, ok := stringValidators[validator]
		if !ok {
			continue
		}

		if err := fn(fieldValue.String(), parameter); err != nil {
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func checkIntField(fieldName string, fieldValue reflect.Value, tags []string) ValidationErrors {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			errs = append(errs, ValidationError{
				Field: fieldName,
				Err:   ErrFieldTagEmptyValue,
			})
			continue
		}

		fn, ok := intValidators[validator]
		if !ok {
			continue
		}

		if err := fn(fieldValue.Int(), parameter); err != nil {
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func checkSliceField(fieldName string, fieldValue reflect.Value, tags []string) ValidationErrors {
	errs := make(ValidationErrors, 0)

	for i := 0; i < fieldValue.Len(); i++ {
		//nolint:exhaustive
		switch fieldValue.Index(i).Kind() {
		case reflect.Int:
			sliceItemErrors := checkIntField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
			errs = append(errs, sliceItemErrors...)
		case reflect.String:
			sliceItemErrors := checkStringField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
			errs = append(errs, sliceItemErrors...)
		default:
			continue
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type ValidationError struct {
	Field string
	Err   error
}

type ValidationErrors []ValidationError

func (v ValidationErrors) Error() string {
	parts := make([]string, len(v))
	for i, v := range v {
		parts[i] = fmt.Sprintf("%s: %s", v.Field, v.Err.Error())
	}
	return strings.Join(parts, "; ")
}

func Validate(v interface{}) error {
	validationErrors := make(ValidationErrors, 0)

	reflectStructWrap := reflect.ValueOf(v)
	if reflectStructWrap.Kind() != reflect.Struct {
		return ErrArgumentNotStructure
	}

	numFields := reflectStructWrap.NumField()

	for i := 0; i < numFields; i++ {
		fieldValue := reflectStructWrap.Field(i)

		if reflectStructWrap.Type().Field(i).Tag == "" {
			continue
		}

		var fieldTags []string
		if tags, ok := reflectStructWrap.Type().Field(i).Tag.Lookup("validate"); ok {
			fieldTags = strings.Split(tags, "|")
		}

		//nolint:exhaustive
		switch fieldValue.Kind() {
		case reflect.String:
			stringFieldErrors := checkStringField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if stringFieldErrors != nil {
				validationErrors = append(validationErrors, stringFieldErrors...)
			}
		case reflect.Int:
			intFieldErrors := checkIntField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if intFieldErrors != nil {
				validationErrors = append(validationErrors, intFieldErrors...)
			}
		case reflect.Slice:
			sliceFieldErrors := checkSliceField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if sliceFieldErrors != nil {
				validationErrors = append(validationErrors, sliceFieldErrors...)
			}
		default:
			continue
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}
