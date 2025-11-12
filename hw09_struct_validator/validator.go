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
	ErrInvalidValidatorParam    = errors.New("invalid validator parameter")
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
			return ErrInvalidValidatorParam, true
		}
		if len(value.String()) != length {
			return ErrFieldStringInvalidLength, false
		}
		return nil, false
	},
	"regexp": func(value reflect.Value, param string) (error, bool) {
		matched, err := regexp.MatchString(param, value.String())
		if err != nil {
			return ErrInvalidValidatorParam, true
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
			return ErrInvalidValidatorParam, true
		}
		if value.Int() > int64(maxValue) {
			return ErrFieldIntBiggerThanMax, false
		}
		return nil, false
	},
	"min": func(value reflect.Value, param string) (error, bool) {
		minValue, err := strconv.Atoi(param)
		if err != nil {
			return ErrInvalidValidatorParam, true
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
				return ErrInvalidValidatorParam, true
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

func checkStringField(fieldName string, fieldValue reflect.Value, tags []string) (ValidationErrors, error) {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			return nil, err
		}

		fn, ok := stringValidators[validator]
		if !ok {
			return nil, err
		}

		if err, internal := fn(fieldValue, parameter); err != nil {
			if internal {
				return nil, err
			}
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs, nil
	}
	return nil, nil
}

func checkIntField(fieldName string, fieldValue reflect.Value, tags []string) (ValidationErrors, error) {
	errs := make(ValidationErrors, 0)

	for _, tag := range tags {
		validator, parameter, err := parseTag(tag)
		if err != nil {
			return nil, err
		}

		fn, ok := intValidators[validator]
		if !ok {
			return nil, err
		}

		if err, internal := fn(fieldValue, parameter); err != nil {
			if internal {
				return nil, err
			}
			errs = append(errs, ValidationError{Field: fieldName, Err: err})
		}
	}

	if len(errs) > 0 {
		return errs, nil
	}
	return nil, nil
}

func checkSliceField(fieldName string, fieldValue reflect.Value, tags []string) (ValidationErrors, error) {
	errs := make(ValidationErrors, 0)

	for i := 0; i < fieldValue.Len(); i++ {
		//nolint:exhaustive
		switch fieldValue.Index(i).Kind() {
		case reflect.Int:
			sliceItemErrors, err := checkIntField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
			if err != nil {
				return nil, err
			}
			errs = append(errs, sliceItemErrors...)
		case reflect.String:
			sliceItemErrors, err := checkStringField(fieldName+fmt.Sprintf("[%d]", i), fieldValue.Index(i), tags)
			if err != nil {
				return nil, err
			}
			errs = append(errs, sliceItemErrors...)
		default:
			return nil, ErrFieldUnsupportedKind
		}
	}

	if len(errs) > 0 {
		return errs, nil
	}
	return nil, nil
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
			stringFieldErrors, err := checkStringField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if err != nil {
				return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: err}
			}
			if stringFieldErrors != nil {
				validationErrors = append(validationErrors, stringFieldErrors...)
			}
		case reflect.Int:
			intFieldErrors, err := checkIntField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if err != nil {
				return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: err}
			}
			if intFieldErrors != nil {
				validationErrors = append(validationErrors, intFieldErrors...)
			}
		case reflect.Slice:
			sliceFieldErrors, err := checkSliceField(reflectStructWrap.Type().Field(i).Name, fieldValue, fieldTags)
			if err != nil {
				return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: err}
			}
			if sliceFieldErrors != nil {
				validationErrors = append(validationErrors, sliceFieldErrors...)
			}
		default:
			return InternalError{Field: reflectStructWrap.Type().Field(i).Name, Err: ErrFieldUnsupportedKind}
		}
	}

	if len(validationErrors) > 0 {
		return validationErrors
	}
	return nil
}
