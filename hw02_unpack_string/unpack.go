package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(input string) (string, error) {
	if input == "" {
		return "", nil
	}

	runes := []rune(input)
	if _, err := strconv.Atoi(string(runes[0])); err == nil {
		return "", ErrInvalidString
	}

	for index := 0; index < len(runes)-1; index++ {
		if _, err := strconv.Atoi(string(runes[index])); err == nil {
			if _, err := strconv.Atoi(string(runes[index+1])); err == nil {
				return "", ErrInvalidString
			}
			continue
		}
	}

	var builder strings.Builder
	for index := 0; index < len(runes); index++ {
		if index == len(runes)-1 {
			if _, err := strconv.Atoi(string(runes[index])); err != nil {
				builder.WriteRune(runes[index])
			}
			break
		}

		if _, err := strconv.Atoi(string(runes[index])); err == nil {
			continue
		}

		times, err := strconv.Atoi(string(runes[index+1]))
		if err != nil {
			builder.WriteRune(runes[index])
		} else {
			builder.WriteString(strings.Repeat(string(runes[index]), times))
		}
	}

	return builder.String(), nil
}
