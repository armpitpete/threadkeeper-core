package strictjson

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

type Code string

const (
	InvalidJSON      Code = "INVALID_JSON"
	DuplicateMember  Code = "DUPLICATE_MEMBER"
	InvalidUTF8      Code = "INVALID_UTF8"
	InvalidNumber    Code = "INVALID_NUMBER"
)

type Error struct {
	Code Code
	Path string
	Msg  string
	Err  error
}

func (e *Error) Error() string {
	if e.Path != "" {
		return fmt.Sprintf("%s at %s: %s", e.Code, e.Path, e.Msg)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Msg)
}

func (e *Error) Unwrap() error { return e.Err }

func IsCode(err error, code Code) bool {
	var e *Error
	return errors.As(err, &e) && e.Code == code
}

func Validate(raw []byte) error {
	if !utf8.Valid(raw) {
		return &Error{Code: InvalidUTF8, Msg: "input is not valid UTF-8"}
	}

	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	if err := parseValue(dec, "$", 0); err != nil {
		return err
	}

	if tok, err := dec.Token(); err != io.EOF {
		if err == nil {
			return &Error{Code: InvalidJSON, Msg: fmt.Sprintf("trailing token %v", tok)}
		}
		return &Error{Code: InvalidJSON, Msg: "trailing data", Err: err}
	}
	return nil
}

func Decode(raw []byte) (any, error) {
	if err := Validate(raw); err != nil {
		return nil, err
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, &Error{Code: InvalidJSON, Msg: "decode failed after strict validation", Err: err}
	}
	return v, nil
}

func parseValue(dec *json.Decoder, path string, depth int) error {
	if depth > 512 {
		return &Error{Code: InvalidJSON, Path: path, Msg: "nesting exceeds safety limit"}
	}
	tok, err := dec.Token()
	if err != nil {
		return &Error{Code: InvalidJSON, Path: path, Msg: "invalid JSON token", Err: err}
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			seen := make(map[string]struct{})
			for dec.More() {
				nameTok, err := dec.Token()
				if err != nil {
					return &Error{Code: InvalidJSON, Path: path, Msg: "invalid object member name", Err: err}
				}
				name, ok := nameTok.(string)
				if !ok {
					return &Error{Code: InvalidJSON, Path: path, Msg: "object member name is not a string"}
				}
				memberPath := path + "/" + escapePointer(name)
				if _, exists := seen[name]; exists {
					return &Error{Code: DuplicateMember, Path: memberPath, Msg: fmt.Sprintf("duplicate member %q", name)}
				}
				seen[name] = struct{}{}
				if err := parseValue(dec, memberPath, depth+1); err != nil {
					return err
				}
			}
			closeTok, err := dec.Token()
			if err != nil || closeTok != json.Delim('}') {
				return &Error{Code: InvalidJSON, Path: path, Msg: "unterminated object", Err: err}
			}
			return nil
		case '[':
			idx := 0
			for dec.More() {
				if err := parseValue(dec, fmt.Sprintf("%s/%d", path, idx), depth+1); err != nil {
					return err
				}
				idx++
			}
			closeTok, err := dec.Token()
			if err != nil || closeTok != json.Delim(']') {
				return &Error{Code: InvalidJSON, Path: path, Msg: "unterminated array", Err: err}
			}
			return nil
		default:
			return &Error{Code: InvalidJSON, Path: path, Msg: fmt.Sprintf("unexpected delimiter %q", t)}
		}
	case json.Number:
		if isNegativeZero(string(t)) {
			return &Error{Code: InvalidNumber, Path: path, Msg: "negative zero is forbidden"}
		}
		return nil
	case string, bool, nil:
		return nil
	default:
		return &Error{Code: InvalidJSON, Path: path, Msg: fmt.Sprintf("unexpected token type %T", tok)}
	}
}

func isNegativeZero(s string) bool {
	if !strings.HasPrefix(s, "-") {
		return false
	}
	s = s[1:]
	if i := strings.IndexAny(s, "eE"); i >= 0 {
		s = s[:i]
	}
	if strings.HasPrefix(s, "0.") {
		s = s[2:]
		if s == "" {
			return false
		}
		for _, r := range s {
			if r != '0' {
				return false
			}
		}
		return true
	}
	return s == "0"
}

func escapePointer(s string) string {
	s = strings.ReplaceAll(s, "~", "~0")
	return strings.ReplaceAll(s, "/", "~1")
}
