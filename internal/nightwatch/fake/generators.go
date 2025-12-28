package fake

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Generator holds the data needed for generating fake data.
type Generator struct {
	Firstnames  []string
	Lastnames   []string
	LoremWords  []string
	Cities      []string
	Counties    []string
	Countries   []string
	Streets     []string
	StreetTypes []string
	Postcodes   []string
}

// NewGenerator creates a Generator with the provided data.
func NewGenerator(firstnames, lastnames, loremWords, cities, counties, countries, streets, streetTypes, postcodes []string) *Generator {
	return &Generator{
		Firstnames:  firstnames,
		Lastnames:   lastnames,
		LoremWords:  loremWords,
		Cities:      cities,
		Counties:    counties,
		Countries:   countries,
		Streets:     streets,
		StreetTypes: streetTypes,
		Postcodes:   postcodes,
	}
}

// Name generates a full name (firstname + lastname).
func (g *Generator) Name(rng RNG) (string, error) {
	first, err := g.Firstname(rng)
	if err != nil {
		return "", err
	}
	last, err := g.Lastname(rng)
	if err != nil {
		return "", err
	}
	return first + " " + last, nil
}

// Firstname generates a random first name.
func (g *Generator) Firstname(rng RNG) (string, error) {
	if len(g.Firstnames) == 0 {
		return "", fmt.Errorf("no firstnames loaded")
	}
	idx, err := rng.Intn(len(g.Firstnames))
	if err != nil {
		return "", err
	}
	return g.Firstnames[idx], nil
}

// Lastname generates a random last name.
func (g *Generator) Lastname(rng RNG) (string, error) {
	if len(g.Lastnames) == 0 {
		return "", fmt.Errorf("no lastnames loaded")
	}
	idx, err := rng.Intn(len(g.Lastnames))
	if err != nil {
		return "", err
	}
	return g.Lastnames[idx], nil
}

// Email generates a random email address.
func (g *Generator) Email(rng RNG) (string, error) {
	first, err := g.Firstname(rng)
	if err != nil {
		return "", err
	}
	last, err := g.Lastname(rng)
	if err != nil {
		return "", err
	}

	domains := []string{"example.com", "test.com", "email.com", "mail.com", "sample.org"}
	domainIdx, err := rng.Intn(len(domains))
	if err != nil {
		return "", err
	}

	// Format: firstname.lastname@domain
	email := strings.ToLower(first) + "." + strings.ToLower(last) + "@" + domains[domainIdx]
	return email, nil
}

// Username generates a random username.
func (g *Generator) Username(rng RNG) (string, error) {
	first, err := g.Firstname(rng)
	if err != nil {
		return "", err
	}

	num, err := rng.Intn(9999)
	if err != nil {
		return "", err
	}

	return strings.ToLower(first) + fmt.Sprintf("%d", num+1), nil
}

// Phone generates a random UK-formatted phone number.
func (g *Generator) Phone(rng RNG) (string, error) {
	// UK phone formats: 07XXX XXXXXX (mobile) or 01onal area codes
	// Using mobile format for simplicity: 07XXX XXXXXX
	prefixes := []string{"07700", "07701", "07702", "07703", "07704", "07705",
		"07706", "07707", "07708", "07709", "07710", "07711", "07712"}

	prefixIdx, err := rng.Intn(len(prefixes))
	if err != nil {
		return "", err
	}

	// Generate 6 random digits for the subscriber number
	digits := make([]int, 6)
	for i := 0; i < 6; i++ {
		digits[i], err = rng.Intn(10)
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%s %d%d%d%d%d%d",
		prefixes[prefixIdx], digits[0], digits[1], digits[2], digits[3], digits[4], digits[5]), nil
}

// Address generates a random UK street address.
func (g *Generator) Address(rng RNG) (string, error) {
	if len(g.Streets) == 0 || len(g.StreetTypes) == 0 {
		return "", fmt.Errorf("no street data loaded")
	}

	num, err := rng.Intn(200)
	if err != nil {
		return "", err
	}
	num++

	nameIdx, err := rng.Intn(len(g.Streets))
	if err != nil {
		return "", err
	}

	typeIdx, err := rng.Intn(len(g.StreetTypes))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%d %s %s", num, g.Streets[nameIdx], g.StreetTypes[typeIdx]), nil
}

// City generates a random UK city name.
func (g *Generator) City(rng RNG) (string, error) {
	if len(g.Cities) == 0 {
		return "", fmt.Errorf("no cities loaded")
	}

	idx, err := rng.Intn(len(g.Cities))
	if err != nil {
		return "", err
	}
	return g.Cities[idx], nil
}

// County generates a random UK county name.
func (g *Generator) County(rng RNG) (string, error) {
	if len(g.Counties) == 0 {
		return "", fmt.Errorf("no counties loaded")
	}

	idx, err := rng.Intn(len(g.Counties))
	if err != nil {
		return "", err
	}
	return g.Counties[idx], nil
}

// Country generates a random country name.
func (g *Generator) Country(rng RNG) (string, error) {
	if len(g.Countries) == 0 {
		return "", fmt.Errorf("no countries loaded")
	}

	idx, err := rng.Intn(len(g.Countries))
	if err != nil {
		return "", err
	}
	return g.Countries[idx], nil
}

// Postcode generates a random UK postcode.
func (g *Generator) Postcode(rng RNG) (string, error) {
	if len(g.Postcodes) == 0 {
		return "", fmt.Errorf("no postcodes loaded")
	}

	// UK postcode format: AA9A 9AA or A9A 9AA or A9 9AA or A99 9AA or AA9 9AA or AA99 9AA
	// We'll use: PREFIX + digit + optional digit + space + digit + two letters
	prefixIdx, err := rng.Intn(len(g.Postcodes))
	if err != nil {
		return "", err
	}
	prefix := g.Postcodes[prefixIdx]

	// First part: prefix + 1-2 digits
	digit1, err := rng.Intn(10)
	if err != nil {
		return "", err
	}

	// Sometimes add a second digit or letter
	var firstPart string
	choice, err := rng.Intn(3)
	if err != nil {
		return "", err
	}
	switch choice {
	case 0:
		firstPart = fmt.Sprintf("%s%d", prefix, digit1)
	case 1:
		digit2, err := rng.Intn(10)
		if err != nil {
			return "", err
		}
		firstPart = fmt.Sprintf("%s%d%d", prefix, digit1, digit2)
	case 2:
		letter, err := rng.Intn(26)
		if err != nil {
			return "", err
		}
		firstPart = fmt.Sprintf("%s%d%c", prefix, digit1, 'A'+letter)
	}

	// Second part: digit + two letters
	inwardDigit, err := rng.Intn(10)
	if err != nil {
		return "", err
	}
	letter1, err := rng.Intn(26)
	if err != nil {
		return "", err
	}
	letter2, err := rng.Intn(26)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %d%c%c", firstPart, inwardDigit, 'A'+letter1, 'A'+letter2), nil
}

// Date generates a random date in YYYY-MM-DD format.
// pastDays: max days in the past (0 means today)
// futureDays: max days in the future (0 means today)
// If both are 0, defaults to random date in past 365 days.
func (g *Generator) Date(rng RNG, pastDays, futureDays int) (string, error) {
	now := time.Now()

	if pastDays == 0 && futureDays == 0 {
		pastDays = 365
	}

	totalRange := pastDays + futureDays
	if totalRange == 0 {
		return now.Format("2006-01-02"), nil
	}

	offset, err := rng.Intn(totalRange + 1)
	if err != nil {
		return "", err
	}

	daysOffset := offset - pastDays
	date := now.AddDate(0, 0, daysOffset)
	return date.Format("2006-01-02"), nil
}

// Datetime generates a random RFC3339 timestamp.
func (g *Generator) Datetime(rng RNG) (string, error) {
	now := time.Now()

	// Random date in past 365 days
	dayOffset, err := rng.Intn(365)
	if err != nil {
		return "", err
	}

	// Random time components
	hour, err := rng.Intn(24)
	if err != nil {
		return "", err
	}
	minute, err := rng.Intn(60)
	if err != nil {
		return "", err
	}
	second, err := rng.Intn(60)
	if err != nil {
		return "", err
	}

	date := now.AddDate(0, 0, -dayOffset)
	datetime := time.Date(date.Year(), date.Month(), date.Day(), hour, minute, second, 0, time.UTC)
	return datetime.Format(time.RFC3339), nil
}

// Time generates a random time in HH:MM:SS format.
func (g *Generator) Time(rng RNG) (string, error) {
	hour, err := rng.Intn(24)
	if err != nil {
		return "", err
	}
	minute, err := rng.Intn(60)
	if err != nil {
		return "", err
	}
	second, err := rng.Intn(60)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%02d:%02d:%02d", hour, minute, second), nil
}

// UUID generates a random UUID v4.
func (g *Generator) UUID(rng RNG) (string, error) {
	// Generate 16 random bytes using our RNG for determinism
	var bytes [16]byte
	for i := 0; i < 16; i++ {
		b, err := rng.Intn(256)
		if err != nil {
			return "", err
		}
		bytes[i] = byte(b)
	}

	// Set version 4 (random)
	bytes[6] = (bytes[6] & 0x0f) | 0x40
	// Set variant bits
	bytes[8] = (bytes[8] & 0x3f) | 0x80

	u, err := uuid.FromBytes(bytes[:])
	if err != nil {
		return "", err
	}
	return u.String(), nil
}

// Hex generates a random hexadecimal string.
func (g *Generator) Hex(rng RNG, length int) (string, error) {
	if length <= 0 {
		length = 16
	}
	bytes := make([]byte, (length+1)/2)
	for i := range bytes {
		b, err := rng.Intn(256)
		if err != nil {
			return "", err
		}
		bytes[i] = byte(b)
	}
	hex := fmt.Sprintf("%x", bytes)
	if len(hex) > length {
		hex = hex[:length]
	}
	return hex, nil
}

// Number generates a random integer in [min, max].
func (g *Generator) Number(rng RNG, min, max int) (int, error) {
	if min > max {
		return 0, fmt.Errorf("min cannot be greater than max")
	}
	if min == max {
		return min, nil
	}
	n, err := rng.Intn(max - min + 1)
	if err != nil {
		return 0, err
	}
	return n + min, nil
}

// URL generates a random URL.
func (g *Generator) URL(rng RNG) (string, error) {
	protocols := []string{"http", "https"}
	subdomains := []string{"www", "api", "app", "dev", "test", "staging"}
	domains := []string{"example", "test", "sample", "demo", "mock"}
	tlds := []string{"com", "org", "net", "io", "dev"}

	protoIdx, err := rng.Intn(len(protocols))
	if err != nil {
		return "", err
	}
	subIdx, err := rng.Intn(len(subdomains))
	if err != nil {
		return "", err
	}
	domIdx, err := rng.Intn(len(domains))
	if err != nil {
		return "", err
	}
	tldIdx, err := rng.Intn(len(tlds))
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s://%s.%s.%s", protocols[protoIdx], subdomains[subIdx], domains[domIdx], tlds[tldIdx]), nil
}

// IPv4 generates a random IPv4 address.
func (g *Generator) IPv4(rng RNG) (string, error) {
	octets := make([]int, 4)
	for i := 0; i < 4; i++ {
		o, err := rng.Intn(256)
		if err != nil {
			return "", err
		}
		octets[i] = o
	}
	return fmt.Sprintf("%d.%d.%d.%d", octets[0], octets[1], octets[2], octets[3]), nil
}

// IPv6 generates a random IPv6 address.
func (g *Generator) IPv6(rng RNG) (string, error) {
	groups := make([]string, 8)
	for i := 0; i < 8; i++ {
		g1, err := rng.Intn(65536)
		if err != nil {
			return "", err
		}
		groups[i] = fmt.Sprintf("%04x", g1)
	}
	return strings.Join(groups, ":"), nil
}

// MAC generates a random MAC address.
func (g *Generator) MAC(rng RNG) (string, error) {
	bytes := make([]string, 6)
	for i := 0; i < 6; i++ {
		b, err := rng.Intn(256)
		if err != nil {
			return "", err
		}
		bytes[i] = fmt.Sprintf("%02x", b)
	}
	return strings.Join(bytes, ":"), nil
}

// Word generates a single random word from the lorem vocabulary.
func (g *Generator) Word(rng RNG) (string, error) {
	if len(g.LoremWords) == 0 {
		return "", fmt.Errorf("no lorem words loaded")
	}
	idx, err := rng.Intn(len(g.LoremWords))
	if err != nil {
		return "", err
	}
	return g.LoremWords[idx], nil
}

// Words generates n random words from the lorem vocabulary.
func (g *Generator) Words(rng RNG, n int) (string, error) {
	if n <= 0 {
		n = 1
	}
	words := make([]string, n)
	for i := 0; i < n; i++ {
		w, err := g.Word(rng)
		if err != nil {
			return "", err
		}
		words[i] = w
	}
	return strings.Join(words, " "), nil
}

// Sentence generates a single sentence (5-15 words, capitalized, with period).
func (g *Generator) Sentence(rng RNG) (string, error) {
	numWords, err := rng.Intn(11) // 0-10
	if err != nil {
		return "", err
	}
	numWords += 5 // 5-15 words

	words, err := g.Words(rng, numWords)
	if err != nil {
		return "", err
	}

	// Capitalize first letter
	if len(words) > 0 {
		words = strings.ToUpper(string(words[0])) + words[1:]
	}
	return words + ".", nil
}

// Sentences generates n sentences.
func (g *Generator) Sentences(rng RNG, n int) (string, error) {
	if n <= 0 {
		n = 1
	}
	sentences := make([]string, n)
	for i := 0; i < n; i++ {
		s, err := g.Sentence(rng)
		if err != nil {
			return "", err
		}
		sentences[i] = s
	}
	return strings.Join(sentences, " "), nil
}

// Paragraph generates a single paragraph (4-8 sentences).
func (g *Generator) Paragraph(rng RNG) (string, error) {
	numSentences, err := rng.Intn(5) // 0-4
	if err != nil {
		return "", err
	}
	numSentences += 4 // 4-8 sentences

	return g.Sentences(rng, numSentences)
}

// Paragraphs generates n paragraphs separated by double newlines.
func (g *Generator) Paragraphs(rng RNG, n int) (string, error) {
	if n <= 0 {
		n = 1
	}
	paragraphs := make([]string, n)
	for i := 0; i < n; i++ {
		p, err := g.Paragraph(rng)
		if err != nil {
			return "", err
		}
		paragraphs[i] = p
	}
	return strings.Join(paragraphs, "\n\n"), nil
}

// Lorem generates lorem ipsum text.
// By default (all args 0), generates 1 paragraph starting with "Lorem ipsum".
func (g *Generator) Lorem(rng RNG, words, sentences, paragraphs int) (string, error) {
	// If no specific count given, default to 1 paragraph
	if words == 0 && sentences == 0 && paragraphs == 0 {
		paragraphs = 1
	}

	var result string
	var err error

	if words > 0 {
		result, err = g.Words(rng, words)
	} else if sentences > 0 {
		result, err = g.Sentences(rng, sentences)
	} else {
		result, err = g.Paragraphs(rng, paragraphs)
	}

	if err != nil {
		return "", err
	}

	// Ensure it starts with "Lorem ipsum" for full paragraphs
	if paragraphs > 0 && !strings.HasPrefix(result, "Lorem ipsum") {
		// Replace the first two words
		parts := strings.SplitN(result, " ", 3)
		if len(parts) >= 3 {
			result = "Lorem ipsum " + parts[2]
		} else if len(parts) >= 1 {
			result = "Lorem ipsum " + strings.Join(parts, " ")
		}
	}

	return result, nil
}
