package hw09structvalidator

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"testing"
)

type UserRole string

// Test the function on different structures and other types.
type (
	User struct {
		ID     string `json:"id" validate:"len:36"`
		Name   string
		Age    int             `validate:"min:18|max:50"`
		Email  string          `validate:"regexp:^\\w+@\\w+\\.\\w+$"`
		Role   UserRole        `validate:"in:admin,stuff"`
		Phones []string        `validate:"len:11"`
		meta   json.RawMessage //nolint:nolintlint
	}

	App struct {
		Version string `validate:"len:5"`
	}

	Price struct {
		Value int `validate:"min"`
	}

	Role struct {
		ID          string `json:"id" validate:"len:5"`
		Permissions []byte `validate:"len:3"`
	}

	ServerError struct {
		Code int `validate:"in:502,503,5o4"`
	}

	Token struct {
		Header    []byte
		Payload   []byte
		Signature []byte
	}

	Response struct {
		Code int    `validate:"in:200,404,500"`
		Body string `json:"omitempty"`
	}
)

type MultiError []error

func (m MultiError) Error() string {
	if len(m) == 0 {
		return ""
	}
	parts := make([]string, 0, len(m))
	for _, e := range m {
		parts = append(parts, e.Error())
	}
	return strings.Join(parts, "; ")
}

func TestValidate(t *testing.T) {
	tests := []struct {
		in          interface{}
		expectedErr error
	}{
		{
			in: ServerError{
				Code: 504,
			},
			expectedErr: strconv.ErrSyntax,
		},
		{
			in: Role{
				ID:          "fsdfs5",
				Permissions: []byte("fsdf"),
			},
			expectedErr: ErrFieldUnsupportedKind,
		},
		{
			in: Price{
				Value: 10,
			},
			expectedErr: ErrFieldTagEmptyValue,
		},
		{
			in:          123,
			expectedErr: ErrArgumentNotStructure,
		},
		{
			in: Token{
				Header:    []byte("{\"alg\": \"HS256\", \"typ\": \"JWT\"}"),
				Payload:   []byte("{\"name\": \"John Doe\", \"admin\": true}"),
				Signature: []byte("sa054lknflw34nrfasmfewo5r"),
			},
			expectedErr: nil,
		},
		{
			in: Response{
				Code: 201,
				Body: "Resource Group not found",
			},
			expectedErr: MultiError{
				ErrFieldIntNotInSet,
			},
		},
		{
			in: App{
				Version: "deluxe",
			},
			expectedErr: MultiError{
				ErrFieldStringInvalidLength,
			},
		},
		{
			in: App{
				Version: "short",
			},
			expectedErr: nil,
		},
		{
			in: User{
				ID:     "nto015lr0g4wxxirdhjuyeyhxzlfg7hc534f",
				Name:   "user",
				Age:    21,
				Email:  "Admin@p.f.ru",
				Role:   "admin",
				Phones: []string{"41241241241", "2", "1", "79999995099"},
				meta:   []byte("{\"message\":\"test\"}"),
			},
			expectedErr: MultiError{
				ErrFieldStringNotMatchRegex,
				ErrFieldStringInvalidLength,
				ErrFieldStringInvalidLength,
			},
		},
		{
			in: User{
				ID:     "nto015lr0g4wxxirdhjuyeyhxzlfg7hc534fG",
				Name:   "user",
				Age:    4,
				Email:  "user@mail.ru",
				Role:   "stuff",
				Phones: []string{"1", "2", "1", "79999995099"},
				meta:   []byte("{\"message\":\"test\"}"),
			},
			expectedErr: MultiError{
				ErrFieldStringInvalidLength,
				ErrFieldIntLessThanMin,
				ErrFieldStringInvalidLength,
				ErrFieldStringInvalidLength,
				ErrFieldStringInvalidLength,
			},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("case %d", i), func(t *testing.T) {
			tt := tt
			t.Parallel()
			checkValidationResult(t, tt.in, tt.expectedErr)
		})
	}
}

func checkValidationResult(t *testing.T, in interface{}, expectedErr error) {
	t.Helper()

	err := Validate(in)

	if expectedErr == nil {
		checkNoError(t, err)
		return
	}
	if err == nil {
		t.Fatalf("expected error %v, got nil", expectedErr)
	}

	var iErr InternalError
	if errors.As(err, &iErr) {
		if !errors.Is(iErr.Err, expectedErr) {
			t.Fatalf("expected internal error %v, got %v", expectedErr, iErr.Err)
		}
		fmt.Println(err)
		return
	}

	var gotErrs ValidationErrors
	if !errors.As(err, &gotErrs) {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}

	var wantErrs MultiError
	if !errors.As(expectedErr, &wantErrs) {
		t.Fatalf("expected MultiError, got %T", expectedErr)
	}

	if len(gotErrs) != len(wantErrs) {
		t.Fatalf("expected %d errors, got %d (%v)", len(wantErrs), len(gotErrs), gotErrs)
	}

	for i := range gotErrs {
		if !errors.Is(gotErrs[i].Err, wantErrs[i]) {
			t.Fatalf("at index %d: expected %v, got %v", i, wantErrs[i], gotErrs[i].Err)
		}
	}
}

func checkNoError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
