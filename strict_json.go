package jep

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"unicode/utf8"
)

// validateEventJSON checks the original bytes before encoding/json can merge
// duplicate members or replace malformed Unicode. It does not verify signatures.
// Number tokens remain untouched for the explicitly selected API profile.
func validateEventJSON(data []byte) error {
	if !utf8.Valid(data) {
		return fmt.Errorf("event contains invalid UTF-8")
	}
	if !json.Valid(data) {
		return fmt.Errorf("event is not one complete JSON value")
	}
	if err := validateSurrogates(data); err != nil {
		return err
	}
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '{' {
		return fmt.Errorf("event must be a JSON object")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := checkJSONValue(decoder, true); err != nil {
		return err
	}
	if _, err := decoder.Token(); err != io.EOF {
		return fmt.Errorf("trailing data after JSON event")
	}
	return nil
}

func checkJSONValue(decoder *json.Decoder, root bool) error {
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	switch token {
	case json.Delim('{'):
		seen := make(map[string]bool)
		for decoder.More() {
			keyToken, err := decoder.Token()
			if err != nil {
				return err
			}
			key, ok := keyToken.(string)
			if !ok {
				return fmt.Errorf("JSON member name is not a string")
			}
			if seen[key] {
				return fmt.Errorf("duplicate JSON member: %q", key)
			}
			seen[key] = true
			// Go's struct decoder accepts case-insensitive aliases. Core names
			// are case-sensitive; do not let an unknown alias change a signed field.
			if root {
				for _, name := range []string{"jep", "id", "verb", "who", "when", "what", "aud", "ref", "ext", "ext_crit", "sig"} {
					if key != name && strings.EqualFold(key, name) {
						return fmt.Errorf("case-aliased Core member: %q", key)
					}
				}
			}
			if err := checkJSONValue(decoder, false); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	case json.Delim('['):
		for decoder.More() {
			if err := checkJSONValue(decoder, false); err != nil {
				return err
			}
		}
		_, err = decoder.Token()
		return err
	}
	return nil
}

// Input is syntactically valid JSON here. Inspect escapes only inside strings;
// a literal escaped backslash followed by "ud800" is not a surrogate escape.
func validateSurrogates(data []byte) error {
	for i := 0; i < len(data); i++ {
		if data[i] != '"' {
			continue
		}
		for i++; i < len(data) && data[i] != '"'; i++ {
			if data[i] != '\\' {
				continue
			}
			i++
			if data[i] != 'u' {
				continue
			}
			code := hexQuad(data[i+1 : i+5])
			i += 4
			if code >= 0xDC00 && code <= 0xDFFF {
				return fmt.Errorf("unpaired low Unicode surrogate")
			}
			if code < 0xD800 || code > 0xDBFF {
				continue
			}
			if i+6 >= len(data) || data[i+1] != '\\' || data[i+2] != 'u' {
				return fmt.Errorf("unpaired high Unicode surrogate")
			}
			low := hexQuad(data[i+3 : i+7])
			if low < 0xDC00 || low > 0xDFFF {
				return fmt.Errorf("invalid Unicode surrogate pair")
			}
			i += 6
		}
	}
	return nil
}

func hexQuad(data []byte) uint16 {
	var result uint16
	for _, b := range data {
		result <<= 4
		switch {
		case b >= '0' && b <= '9':
			result += uint16(b - '0')
		case b >= 'a' && b <= 'f':
			result += uint16(b-'a') + 10
		case b >= 'A' && b <= 'F':
			result += uint16(b-'A') + 10
		}
	}
	return result
}
