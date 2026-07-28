package models

import (
	"encoding/json"
	"strings"
)

const learningItemPlaceholderMaxLen = 64

// validateLearningItemPlaceholdersInData walks all string leaves inside a block
// data JSON value and validates {{identifier}} placeholder syntax only.
func validateLearningItemPlaceholdersInData(data json.RawMessage) error {
	var root interface{}
	if err := json.Unmarshal(data, &root); err != nil {
		return ErrLearningItemMetadataInvalid
	}
	return walkLearningItemPlaceholderValue(root)
}

func walkLearningItemPlaceholderValue(value interface{}) error {
	switch typed := value.(type) {
	case string:
		return validateLearningItemPlaceholderString(typed)
	case map[string]interface{}:
		for _, child := range typed {
			if err := walkLearningItemPlaceholderValue(child); err != nil {
				return err
			}
		}
	case []interface{}:
		for _, child := range typed {
			if err := walkLearningItemPlaceholderValue(child); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateLearningItemPlaceholderString(text string) error {
	for i := 0; i < len(text); {
		if text[i] != '{' {
			i++
			continue
		}
		if i+1 >= len(text) || text[i+1] != '{' {
			// Lone '{' is ordinary content.
			i++
			continue
		}
		contentStart := i + 2
		j := contentStart
		closed := false
		for j < len(text) {
			if j+1 < len(text) && text[j] == '{' && text[j+1] == '{' {
				return ErrLearningItemPlaceholderSyntax
			}
			if j+1 < len(text) && text[j] == '}' && text[j+1] == '}' {
				content := text[contentStart:j]
				if content == "" {
					return ErrLearningItemPlaceholderSyntax
				}
				// Any brace inside the token body is a delimiter malformation
				// (covers nested forms like {{{name}}}).
				if stringsContainsByte(content, '{') || stringsContainsByte(content, '}') {
					return ErrLearningItemPlaceholderSyntax
				}
				if !isValidLearningItemPlaceholderIdentifier(content) {
					return ErrLearningItemPlaceholderInvalid
				}
				i = j + 2
				closed = true
				break
			}
			j++
		}
		if !closed {
			return ErrLearningItemPlaceholderSyntax
		}
	}
	return nil
}

func isValidLearningItemPlaceholderIdentifier(identifier string) bool {
	if len(identifier) == 0 || len(identifier) > learningItemPlaceholderMaxLen {
		return false
	}
	for i := 0; i < len(identifier); i++ {
		c := identifier[i]
		if i == 0 {
			if !isASCIILetter(c) {
				return false
			}
			continue
		}
		if c == '_' || isASCIILetter(c) || isASCIIDigit(c) {
			continue
		}
		return false
	}
	return true
}

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

func isASCIIDigit(c byte) bool {
	return c >= '0' && c <= '9'
}

func stringsContainsByte(s string, b byte) bool {
	return strings.IndexByte(s, b) >= 0
}
