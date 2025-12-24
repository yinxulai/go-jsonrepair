package jsonrepair

import (
	"fmt"
	"strings"
	"unicode"
)

// Repair repairs a malformed JSON string and returns valid JSON
func Repair(input string) (string, error) {
	p := &parser{
		input:  input,
		index:  0,
		output: strings.Builder{},
	}
	return p.parse()
}

type parser struct {
	input  string
	index  int
	output strings.Builder
}

func (p *parser) parse() (string, error) {
	p.skipWhitespaceAndComments()

	// Check for code fence like ```json ... ```
	if p.peekCodeFence() {
		return p.parseCodeFence()
	}

	// Check for JSONP wrapper like callback({...})
	if p.peekFunc() {
		return p.parseJSONPWrapper()
	}

	if err := p.parseValue(); err != nil {
		return "", err
	}

	p.skipWhitespaceAndComments()

	return p.output.String(), nil
}

func (p *parser) parseValue() error {
	p.skipWhitespaceAndComments()

	if p.index >= len(p.input) {
		return fmt.Errorf("unexpected end of input")
	}

	char := p.input[p.index]

	switch {
	case char == '{':
		return p.parseObject()
	case char == '[':
		return p.parseArray()
	case char == '"':
		return p.parseString()
	case char == '\'':
		return p.parseSingleQuotedString()
	case char == 'n':
		return p.parseKeyword("null")
	case char == 't':
		return p.parseKeyword("true")
	case char == 'f':
		return p.parseKeyword("false")
	case char == 'N':
		// Check for NumberLong, NumberInt, or None
		if p.peekKeyword("NumberLong") {
			p.index += len("NumberLong")
			return p.parseMongoDBType()
		} else if p.peekKeyword("NumberInt") {
			p.index += len("NumberInt")
			return p.parseMongoDBType()
		} else if p.matchKeyword("None") {
			p.output.WriteString("null")
			return nil
		}
		return p.parseUnquotedString()
	case char == 'T':
		// Python True
		if p.matchKeyword("True") {
			p.output.WriteString("true")
			return nil
		}
		return p.parseUnquotedString()
	case char == 'F':
		// Python False
		if p.matchKeyword("False") {
			p.output.WriteString("false")
			return nil
		}
		return p.parseUnquotedString()
	case char == 'I':
		// MongoDB ISODate
		if p.peekKeyword("ISODate") {
			p.index += len("ISODate")
			return p.parseMongoDBType()
		}
		return p.parseUnquotedString()
	case char == 'O':
		// MongoDB ObjectId
		if p.peekKeyword("ObjectId") {
			p.index += len("ObjectId")
			return p.parseMongoDBType()
		}
		return p.parseUnquotedString()
	case char == '-' || (char >= '0' && char <= '9'):
		return p.parseNumber()
	case unicode.IsLetter(rune(char)) || char == '_' || char == '$':
		// Unquoted string (likely an unquoted key or special value)
		return p.parseUnquotedString()
	default:
		return fmt.Errorf("unexpected character '%c' at position %d", char, p.index)
	}
}

func (p *parser) parseObject() error {
	p.output.WriteByte('{')
	p.index++ // skip '{'
	p.skipWhitespaceAndComments()

	first := true
	for p.index < len(p.input) && p.input[p.index] != '}' {
		if !first {
			p.output.WriteByte(',')
		}
		first = false

		p.skipWhitespaceAndComments()

		// Parse key
		if err := p.parseKey(); err != nil {
			return err
		}

		p.skipWhitespaceAndComments()

		// Expect colon
		if p.index >= len(p.input) {
			// Truncated - add closing brace
			p.output.WriteByte('}')
			return nil
		}

		if p.input[p.index] != ':' {
			return fmt.Errorf("expected ':' at position %d", p.index)
		}
		p.output.WriteByte(':')
		p.index++

		p.skipWhitespaceAndComments()

		// Parse value
		if p.index >= len(p.input) {
			// Truncated - add null and close
			p.output.WriteString("null}")
			return nil
		}

		if err := p.parseValue(); err != nil {
			return err
		}

		p.skipWhitespaceAndComments()

		// Check for comma or end
		if p.index < len(p.input) && p.input[p.index] == ',' {
			p.index++
			p.skipWhitespaceAndComments()
			// Check for trailing comma
			if p.index < len(p.input) && p.input[p.index] == '}' {
				// Skip the comma we just saw, don't output it
				break
			}
		}
	}

	if p.index >= len(p.input) {
		// Truncated - close the object
		p.output.WriteByte('}')
		return nil
	}

	p.output.WriteByte('}')
	p.index++ // skip '}'
	return nil
}

func (p *parser) parseArray() error {
	p.output.WriteByte('[')
	p.index++ // skip '['
	p.skipWhitespaceAndComments()

	first := true
	for p.index < len(p.input) && p.input[p.index] != ']' {
		p.skipWhitespaceAndComments()

		// Check for ellipsis (...) and skip it
		if p.index+2 < len(p.input) && p.input[p.index:p.index+3] == "..." {
			p.index += 3
			p.skipWhitespaceAndComments()
			// Skip comma after ellipsis if present
			if p.index < len(p.input) && p.input[p.index] == ',' {
				p.index++
				p.skipWhitespaceAndComments()
			}
			// Check if array ends after ellipsis
			if p.index >= len(p.input) || p.input[p.index] == ']' {
				break
			}
			// Continue to parse next value
		}

		if !first {
			p.output.WriteByte(',')
		}
		first = false

		if err := p.parseValue(); err != nil {
			return err
		}

		p.skipWhitespaceAndComments()

		// Check for comma or end
		if p.index < len(p.input) && p.input[p.index] == ',' {
			p.index++
			p.skipWhitespaceAndComments()
			// Check for trailing comma or ellipsis
			if p.index < len(p.input) {
				if p.input[p.index] == ']' {
					break
				}
				// Check for ellipsis after comma
				if p.index+2 < len(p.input) && p.input[p.index:p.index+3] == "..." {
					p.index += 3
					p.skipWhitespaceAndComments()
					// Check if more values follow
					if p.index >= len(p.input) || p.input[p.index] == ']' {
						break
					}
					// Skip comma after ellipsis if present
					if p.input[p.index] == ',' {
						p.index++
						p.skipWhitespaceAndComments()
					}
				}
			}
		}
	}

	if p.index >= len(p.input) {
		// Truncated - close the array
		p.output.WriteByte(']')
		return nil
	}

	p.output.WriteByte(']')
	p.index++ // skip ']'
	return nil
}

func (p *parser) parseKey() error {
	if p.index >= len(p.input) {
		return fmt.Errorf("unexpected end of input while parsing key")
	}

	char := p.input[p.index]

	if char == '"' {
		return p.parseString()
	} else if char == '\'' {
		return p.parseSingleQuotedString()
	} else {
		// Unquoted key - need to add quotes
		return p.parseUnquotedKey()
	}
}

func (p *parser) parseString() error {
	p.output.WriteByte('"')
	p.index++ // skip opening quote

	for p.index < len(p.input) {
		char := p.input[p.index]

		if char == '"' {
			p.index++

			// Check for concatenation with +
			savedIndex := p.index
			p.skipWhitespaceAndComments()
			if p.index < len(p.input) && p.input[p.index] == '+' {
				p.index++
				p.skipWhitespaceAndComments()
				if p.index < len(p.input) && (p.input[p.index] == '"' || p.input[p.index] == '\'') {
					// Continue concatenating - don't close the quote yet; skip opening quote of next string
					p.index++ // skip opening quote (single or double)
					continue
				}
			}
			// No concatenation, restore index and close quote
			p.index = savedIndex
			p.output.WriteByte('"')
			return nil
		} else if char == '\\' {
			if p.index+1 < len(p.input) {
				p.output.WriteByte('\\')
				p.index++
				p.output.WriteByte(p.input[p.index])
				p.index++
			} else {
				p.output.WriteByte('\\')
				p.index++
			}
		} else {
			p.output.WriteByte(char)
			p.index++
		}
	}

	// Unterminated string - close it
	p.output.WriteByte('"')
	return nil
}

func (p *parser) parseSingleQuotedString() error {
	p.output.WriteByte('"') // Convert to double quote
	p.index++               // skip opening single quote

	for p.index < len(p.input) {
		char := p.input[p.index]

		if char == '\'' {
			p.index++

			// Check for concatenation
			savedIndex := p.index
			p.skipWhitespaceAndComments()
			if p.index < len(p.input) && p.input[p.index] == '+' {
				p.index++
				p.skipWhitespaceAndComments()
				if p.index < len(p.input) && (p.input[p.index] == '"' || p.input[p.index] == '\'') {
					// Continue concatenating
					p.index++
					continue
				}
			}
			// No concatenation, restore and close
			p.index = savedIndex
			p.output.WriteByte('"') // Convert to double quote
			return nil
		} else if char == '\\' {
			if p.index+1 < len(p.input) {
				nextChar := p.input[p.index+1]
				if nextChar == '\'' {
					// Escaped single quote - just output the quote
					p.output.WriteByte('\'')
					p.index += 2
				} else {
					p.output.WriteByte('\\')
					p.index++
					p.output.WriteByte(p.input[p.index])
					p.index++
				}
			} else {
				p.output.WriteByte('\\')
				p.index++
			}
		} else if char == '"' {
			// Double quote inside single-quoted string needs to be escaped
			p.output.WriteString("\\\"")
			p.index++
		} else {
			p.output.WriteByte(char)
			p.index++
		}
	}

	// Unterminated string - close it
	p.output.WriteByte('"')
	return nil
}

func (p *parser) parseUnquotedKey() error {
	start := p.index

	// Read until we hit a colon, whitespace, or comment
	for p.index < len(p.input) {
		char := p.input[p.index]
		if char == ':' || unicode.IsSpace(rune(char)) || char == '/' {
			break
		}
		p.index++
	}

	key := p.input[start:p.index]
	p.output.WriteByte('"')
	p.output.WriteString(key)
	p.output.WriteByte('"')

	return nil
}

func (p *parser) parseUnquotedString() error {
	// This handles unquoted strings that should be quoted
	// We quote them as strings
	start := p.index

	for p.index < len(p.input) {
		char := p.input[p.index]
		if unicode.IsSpace(rune(char)) || char == ',' || char == '}' || char == ']' || char == ':' {
			break
		}
		p.index++
	}

	value := p.input[start:p.index]
	p.output.WriteByte('"')
	p.output.WriteString(value)
	p.output.WriteByte('"')

	return nil
}

func (p *parser) parseNumber() error {
	start := p.index

	// Optional minus
	if p.index < len(p.input) && p.input[p.index] == '-' {
		p.index++
	}

	// Integer part
	if p.index >= len(p.input) {
		return fmt.Errorf("invalid number at position %d", start)
	}

	if p.input[p.index] == '0' {
		p.index++
	} else if p.input[p.index] >= '1' && p.input[p.index] <= '9' {
		for p.index < len(p.input) && p.input[p.index] >= '0' && p.input[p.index] <= '9' {
			p.index++
		}
	} else {
		return fmt.Errorf("invalid number at position %d", start)
	}

	// Fractional part
	if p.index < len(p.input) && p.input[p.index] == '.' {
		p.index++
		if p.index >= len(p.input) || p.input[p.index] < '0' || p.input[p.index] > '9' {
			return fmt.Errorf("invalid number at position %d", start)
		}
		for p.index < len(p.input) && p.input[p.index] >= '0' && p.input[p.index] <= '9' {
			p.index++
		}
	}

	// Exponent part
	if p.index < len(p.input) && (p.input[p.index] == 'e' || p.input[p.index] == 'E') {
		p.index++
		if p.index < len(p.input) && (p.input[p.index] == '+' || p.input[p.index] == '-') {
			p.index++
		}
		if p.index >= len(p.input) || p.input[p.index] < '0' || p.input[p.index] > '9' {
			return fmt.Errorf("invalid number at position %d", start)
		}
		for p.index < len(p.input) && p.input[p.index] >= '0' && p.input[p.index] <= '9' {
			p.index++
		}
	}

	p.output.WriteString(p.input[start:p.index])
	return nil
}

func (p *parser) parseKeyword(keyword string) error {
	if !p.matchKeyword(keyword) {
		return fmt.Errorf("expected '%s' at position %d", keyword, p.index)
	}
	p.output.WriteString(keyword)
	return nil
}

func (p *parser) matchKeyword(keyword string) bool {
	if p.index+len(keyword) > len(p.input) {
		return false
	}
	if p.input[p.index:p.index+len(keyword)] == keyword {
		p.index += len(keyword)
		return true
	}
	return false
}

func (p *parser) peekKeyword(keyword string) bool {
	if p.index+len(keyword) > len(p.input) {
		return false
	}
	return p.input[p.index:p.index+len(keyword)] == keyword
}

func (p *parser) parseMongoDBType() error {
	// Function name already consumed by caller
	p.skipWhitespaceAndComments()

	if p.index >= len(p.input) || p.input[p.index] != '(' {
		return fmt.Errorf("invalid MongoDB type - expected '(' at position %d", p.index)
	}

	p.index++ // skip '('
	p.skipWhitespaceAndComments()

	// Parse the content
	if err := p.parseValue(); err != nil {
		return err
	}

	p.skipWhitespaceAndComments()

	if p.index >= len(p.input) {
		return fmt.Errorf("invalid MongoDB type - expected ')' but reached end of input")
	}
	if p.input[p.index] != ')' {
		return fmt.Errorf("invalid MongoDB type - expected ')' but found '%c' at position %d", p.input[p.index], p.index)
	}

	p.index++ // skip ')'

	return nil
}

func (p *parser) skipWhitespaceAndComments() {
	for p.index < len(p.input) {
		char := p.input[p.index]

		if unicode.IsSpace(rune(char)) {
			p.index++
		} else if char == '/' && p.index+1 < len(p.input) {
			nextChar := p.input[p.index+1]
			if nextChar == '/' {
				// Single-line comment
				p.index += 2
				for p.index < len(p.input) && p.input[p.index] != '\n' {
					p.index++
				}
			} else if nextChar == '*' {
				// Multi-line comment
				p.index += 2
				for p.index+1 < len(p.input) {
					if p.input[p.index] == '*' && p.input[p.index+1] == '/' {
						p.index += 2
						break
					}
					p.index++
				}
			} else {
				break
			}
		} else {
			break
		}
	}
}

func (p *parser) peekFunc() bool {
	// Check if this looks like a function call (JSONP wrapper)
	saved := p.index

	// Skip identifier
	if p.index >= len(p.input) || !unicode.IsLetter(rune(p.input[p.index])) {
		return false
	}

	for p.index < len(p.input) && (unicode.IsLetter(rune(p.input[p.index])) || unicode.IsDigit(rune(p.input[p.index])) || p.input[p.index] == '_') {
		p.index++
	}

	// Skip whitespace
	for p.index < len(p.input) && unicode.IsSpace(rune(p.input[p.index])) {
		p.index++
	}

	// Check for opening parenthesis
	result := p.index < len(p.input) && p.input[p.index] == '('
	p.index = saved
	return result
}

func (p *parser) peekCodeFence() bool {
	if p.index+2 < len(p.input) && p.input[p.index:p.index+3] == "```" {
		return true
	}
	return false
}

func (p *parser) parseJSONPWrapper() (string, error) {
	// Skip function name
	for p.index < len(p.input) && (unicode.IsLetter(rune(p.input[p.index])) || unicode.IsDigit(rune(p.input[p.index])) || p.input[p.index] == '_') {
		p.index++
	}

	p.skipWhitespaceAndComments()

	if p.index >= len(p.input) || p.input[p.index] != '(' {
		return "", fmt.Errorf("expected '(' for JSONP wrapper")
	}
	p.index++ // skip '('

	p.skipWhitespaceAndComments()

	// Now parse the actual JSON value
	if err := p.parseValue(); err != nil {
		return "", err
	}

	p.skipWhitespaceAndComments()

	// Skip closing parenthesis if present
	if p.index < len(p.input) && p.input[p.index] == ')' {
		p.index++
	}

	return p.output.String(), nil
}

func (p *parser) parseCodeFence() (string, error) {
	// Skip opening ```
	p.index += 3

	// Skip optional language identifier (e.g., "json").
	// Stop early if we encounter characters that look like the start of JSON
	// content, to avoid skipping over the JSON when there is no newline.
	for p.index < len(p.input) && p.input[p.index] != '\n' {
		ch := p.input[p.index]
		// Check for characters that definitely start JSON values
		if ch == '{' || ch == '[' || ch == '"' || 
		   (ch >= '0' && ch <= '9') || ch == '-' {
			break
		}
		// For single quotes, true, false, null - check if followed by valid JSON context
		// This is a heuristic: if we see these after whitespace, it's likely JSON
		if (ch == '\'' || ch == 't' || ch == 'f' || ch == 'n') && p.index > 3 {
			// Check if there's whitespace before this character
			prevChar := p.input[p.index-1]
			if prevChar == ' ' || prevChar == '\t' {
				break
			}
		}
		p.index++
	}
	if p.index < len(p.input) && p.input[p.index] == '\n' {
		p.index++ // skip newline
	}

	p.skipWhitespaceAndComments()

	// Parse the JSON content
	if err := p.parseValue(); err != nil {
		return "", err
	}

	p.skipWhitespaceAndComments()

	// Skip closing ```
	if p.index+2 < len(p.input) && p.input[p.index:p.index+3] == "```" {
		p.index += 3
	}

	return p.output.String(), nil
}
