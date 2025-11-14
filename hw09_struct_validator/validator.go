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

type validatorFunc func(value reflect.Value, param string) (err error, internal bool)

var stringValidators = map[string]validatorFunc{
	"len": func(value reflect.Value, param string) (error, bool) {
		length, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("invalid len param %q: %w", param, err), true
		}
		if len(value.String()) != length {
			return ErrFieldStringInvalidLength, false
		}
		return nil, false
	},
	"regexp": func(value reflect.Value, param string) (error, bool) {
		matched, err := regexp.MatchString(param, value.String())
		if err != nil {
			return fmt.Errorf("invalid regexp param %q: %w", param, err), true
		}
		if !matched {
			return ErrFieldStringNotMatchRegex, false
		}
		return nil, false
	},
	"in": func(value reflect.Value, param string) (error, bool) {
		set := strings.Split(param, ",")
		for _, v := range set {
			if value.String() == v {
				return nil, false
			}
		}
		return ErrFieldStringNotInSet, false
	},
}

var intValidators = map[string]validatorFunc{
	"max": func(value reflect.Value, param string) (error, bool) {
		maxValue, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("invalid maxValue param %q: %w", param, err), true
		}
		if value.Int() > int64(maxValue) {
			return ErrFieldIntBiggerThanMax, false
		}
		return nil, false
	},
	"min": func(value reflect.Value, param string) (error, bool) {
		minValue, err := strconv.Atoi(param)
		if err != nil {
			return fmt.Errorf("invalid minValue param %q: %w", param, err), true
		}
		if value.Int() < int64(minValue) {
			return ErrFieldIntLessThanMin, false
		}
		return nil, false
	},
	"in": func(value reflect.Value, param string) (error, bool) {
		set := strings.Split(param, ",")
		found := false
		for _, part := range set {
			val, err := strconv.ParseInt(part, 10, 64)
			if err != nil {
				return fmt.Errorf("invalid param %q in set: %w", param, err), true
			}
			if value.Int() == val {
				found = true
				break
			}
		}
		if !found {
			return ErrFieldIntNotInSet, false
		}
		return nil, false
	},
}

func parseTag(tag string) (string, string, error) {
	tagParts := strings.SplitN(tag, ":", 2)
	if len(tagParts) < 2 {
		return "", "", ErrFieldTagEmptyValue
	}
	return tagParts[0], tagParts[1], nil
}

func checkStringField(fieldName string, fieldValue reflect.Value, tags []string) error {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			return fmt.Errorf("parse tag: %w", err)
		}

		fn, ok := stringValidators[validator]
		if !ok {
			return fmt.Errorf("invalid validator %q: %w", validator, err)
		}

		if err, internal := fn(fieldValue, parameter); err != nil {
			if internal {
				return err
			}
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func checkIntField(fieldName string, fieldValue reflect.Value, tags []string) error {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			return fmt.Errorf("parse tag: %w", err)
		}

		fn, ok := intValidators[validator]
		if !ok {
			return fmt.Errorf("invalid validator %q: %w", validator, err)
		}

		if err, internal := fn(fieldValue, parameter); err != nil {
			if internal {
				return err
			}
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func checkSliceField(fieldName string, fieldValue reflect.Value, tags []string) error {
	errs := make(ValidationErrors, 0)

	for i := 0; i < fieldValue.Len(); i++ {
		var err error
		//nolint:exhaustive
		switch fieldValue.Index(i).Kind() {
		case reflect.Int:
			err = checkIntField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
		case reflect.String:
			err = checkStringField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
		default:
			return fmt.Errorf("unsupported slice element kind %v: %w", fieldValue.Index(i).Kind(), ErrFieldUnsupportedKind)
		}

		if err == nil {
			continue
		}

		var vErrs ValidationErrors
		if errors.As(err, &vErrs) {
			errs = append(errs, vErrs...)
			continue
		}

		return err
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

type InternalError struct {
	Field string
	Err   error
}

func (e InternalError) Error() string {
	return fmt.Sprintf("internal error on field %q: %v", e.Field, e.Err)
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
		return InternalError{Field: "", Err: ErrArgumentNotStructure}
	}

	numFields := reflectStructWrap.NumField()

	for i := 0; i < numFields; i++ {
		fieldValue := reflectStructWrap.Field(i)

		tags, ok := reflectStructWrap.Type().Field(i).Tag.Lookup("validate")
		if !ok {
			continue
		}
		fieldTags := strings.Split(tags, "|")

		var err error
		//nolint:exhaustive
		switch fieldValue.Kind() {
		case reflect.String:
			err = checkStringField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
		case reflect.Int:
			err = checkIntField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
		case reflect.Slice:
			err = checkSliceField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
		default:
			return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: ErrFieldUnsupportedKind}
		}

		if err == nil {
			continue
		}

		var vErrs ValidationErrors
		if errors.As(err, &vErrs) {
			validationErrors = append(validationErrors, vErrs...)
			continue
		}
		return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: err}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}
