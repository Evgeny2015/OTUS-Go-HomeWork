package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

type ParsedRune struct {
	Sym rune
	Cnt int
}

func Unpack(str string) (string, error) {
	if str == "" {
		return "", nil
	}

	parsed := []ParsedRune{}
	var lastRune rune
	mark := false

	for _, s := range str {
		if mark {
			lastRune = s
			mark = false
			continue
		}

		if unicode.IsDigit(s) {
			if lastRune == 0 {
				return "", ErrInvalidString
			}

			cnt, _ := strconv.Atoi(string(s))
			parsed = append(parsed, ParsedRune{Sym: lastRune, Cnt: cnt})
			lastRune = 0
			continue
		}

		if s == '\\' {
			mark = true
		}

		// all others - symbols
		if lastRune != 0 {
			parsed = append(parsed, ParsedRune{Sym: lastRune, Cnt: 1})
		}
		lastRune = s
	}

	// append last symbol if exists
	if lastRune != 0 {
		parsed = append(parsed, ParsedRune{Sym: lastRune, Cnt: 1})
	}

	// build string
	var builder strings.Builder
	for _, p := range parsed {
		builder.WriteString(strings.Repeat(string(p.Sym), p.Cnt))
	}

	return builder.String(), nil
}
