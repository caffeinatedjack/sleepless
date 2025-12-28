package fake

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Token represents a parsed template token.
type Token struct {
	Literal     string   // literal text (if IsLiteral)
	Type        string   // generator type (if !IsLiteral)
	Args        []string // arguments for the generator
	IsLiteral   bool
	RawTemplate string // the original {{...}} string for error messages
}

// Template represents a parsed template.
type Template struct {
	Raw    string
	Tokens []Token
}

// templateRegex matches {{type}} or {{type:arg1:arg2:...}}
var templateRegex = regexp.MustCompile(`\{\{(\w+)(?::([^}]+))?\}\}`)

// ParseTemplate parses a template string into tokens.
func ParseTemplate(s string) (*Template, error) {
	t := &Template{Raw: s}

	matches := templateRegex.FindAllStringSubmatchIndex(s, -1)
	if len(matches) == 0 {
		// No placeholders, entire string is literal
		t.Tokens = []Token{{Literal: s, IsLiteral: true}}
		return t, nil
	}

	lastEnd := 0
	for _, match := range matches {
		// match[0], match[1] = full match start/end
		// match[2], match[3] = type group start/end
		// match[4], match[5] = args group start/end (if present)

		// Add literal before this match
		if match[0] > lastEnd {
			t.Tokens = append(t.Tokens, Token{
				Literal:   s[lastEnd:match[0]],
				IsLiteral: true,
			})
		}

		// Extract type
		typeName := s[match[2]:match[3]]

		// Extract args if present
		var args []string
		if match[4] != -1 {
			argsStr := s[match[4]:match[5]]
			args = strings.Split(argsStr, ":")
		}

		t.Tokens = append(t.Tokens, Token{
			Type:        typeName,
			Args:        args,
			IsLiteral:   false,
			RawTemplate: s[match[0]:match[1]],
		})

		lastEnd = match[1]
	}

	// Add any remaining literal after the last match
	if lastEnd < len(s) {
		t.Tokens = append(t.Tokens, Token{
			Literal:   s[lastEnd:],
			IsLiteral: true,
		})
	}

	return t, nil
}

// ValidTypes returns the list of valid generator types.
func ValidTypes() []string {
	return []string{
		"name", "firstname", "lastname", "email", "username", "phone",
		"address", "city", "county", "country", "postcode",
		"date", "datetime", "time",
		"uuid", "hex", "number",
		"lorem", "word", "sentence", "paragraph",
		"url", "ipv4", "ipv6", "mac",
	}
}

// isValidType checks if a type name is valid.
func isValidType(typeName string) bool {
	for _, t := range ValidTypes() {
		if t == typeName {
			return true
		}
	}
	return false
}

// Validate checks that all placeholders use valid types.
func (t *Template) Validate() error {
	for _, token := range t.Tokens {
		if !token.IsLiteral {
			if !isValidType(token.Type) {
				return fmt.Errorf("unknown type %q in template placeholder %s", token.Type, token.RawTemplate)
			}
		}
	}
	return nil
}

// Render renders the template using the generator and RNG.
func (t *Template) Render(g *Generator, rng RNG) (string, error) {
	var result strings.Builder

	for _, token := range t.Tokens {
		if token.IsLiteral {
			result.WriteString(token.Literal)
			continue
		}

		value, err := generateForType(g, rng, token.Type, token.Args)
		if err != nil {
			return "", fmt.Errorf("error generating %s: %w", token.RawTemplate, err)
		}
		result.WriteString(value)
	}

	return result.String(), nil
}

// generateForType generates a value for the given type and args.
func generateForType(g *Generator, rng RNG, typeName string, args []string) (string, error) {
	switch typeName {
	case "name":
		return g.Name(rng)
	case "firstname":
		return g.Firstname(rng)
	case "lastname":
		return g.Lastname(rng)
	case "email":
		return g.Email(rng)
	case "username":
		return g.Username(rng)
	case "phone":
		return g.Phone(rng)
	case "address":
		return g.Address(rng)
	case "city":
		return g.City(rng)
	case "county":
		return g.County(rng)
	case "country":
		return g.Country(rng)
	case "postcode":
		return g.Postcode(rng)
	case "date":
		past, future := 0, 0
		if len(args) >= 2 {
			switch args[0] {
			case "past":
				past, _ = strconv.Atoi(args[1])
			case "future":
				future, _ = strconv.Atoi(args[1])
			}
		}
		return g.Date(rng, past, future)
	case "datetime":
		return g.Datetime(rng)
	case "time":
		return g.Time(rng)
	case "uuid":
		return g.UUID(rng)
	case "hex":
		length := 16
		if len(args) >= 1 {
			length, _ = strconv.Atoi(args[0])
		}
		return g.Hex(rng, length)
	case "number":
		min, max := 0, 100
		if len(args) >= 2 {
			min, _ = strconv.Atoi(args[0])
			max, _ = strconv.Atoi(args[1])
		} else if len(args) == 1 {
			max, _ = strconv.Atoi(args[0])
		}
		n, err := g.Number(rng, min, max)
		if err != nil {
			return "", err
		}
		return strconv.Itoa(n), nil
	case "lorem":
		words, sentences, paragraphs := 0, 0, 0
		if len(args) >= 2 {
			switch args[0] {
			case "words":
				words, _ = strconv.Atoi(args[1])
			case "sentences":
				sentences, _ = strconv.Atoi(args[1])
			case "paragraphs":
				paragraphs, _ = strconv.Atoi(args[1])
			}
		}
		return g.Lorem(rng, words, sentences, paragraphs)
	case "word":
		return g.Word(rng)
	case "sentence":
		return g.Sentence(rng)
	case "paragraph":
		return g.Paragraph(rng)
	case "url":
		return g.URL(rng)
	case "ipv4":
		return g.IPv4(rng)
	case "ipv6":
		return g.IPv6(rng)
	case "mac":
		return g.MAC(rng)
	default:
		return "", fmt.Errorf("unknown type: %s", typeName)
	}
}
