package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func isDigit(r rune) (int, bool) {
	d, err := strconv.Atoi(string(r))
	return d, err == nil
}

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	runes := []rune(input)
	if _, ok := isDigit(runes[0]); ok {
		return "", ErrInvalidString
	}

	var builder strings.Builder
	for index := 0; index < len(runes); index++ {
		_, ok := isDigit(runes[index])

		if index == len(runes)-1 {
			if !ok {
				builder.WriteRune(runes[index])
			}
			break
		}

		if ok {
			if _, ok := isDigit(runes[index+1]); ok {
				return "", ErrInvalidString
			}
			continue
		}

		if times, ok := isDigit(runes[index+1]); !ok {
			builder.WriteRune(runes[index])
		} else {
			builder.WriteString(strings.Repeat(string(runes[index]), times))
		}
	}

	return builder.String(), nil
}
