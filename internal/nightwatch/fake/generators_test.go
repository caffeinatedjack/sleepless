package fake

import (
	"regexp"
	"strings"
	"testing"

	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
)

func testGenerator() *Generator {
	firstnames := []string{"John", "Jane", "Alice", "Bob"}
	lastnames := []string{"Smith", "Doe", "Johnson", "Williams"}
	loremWords := []string{"lorem", "ipsum", "dolor", "sit", "amet", "consectetur", "adipiscing", "elit"}
	cities := []string{"London", "Manchester", "Birmingham", "Leeds"}
	counties := []string{"Greater London", "West Midlands", "Greater Manchester"}
	countries := []string{"United Kingdom", "Ireland", "France", "Germany"}
	streets := []string{"High", "Church", "Station", "Victoria"}
	streetTypes := []string{"Street", "Road", "Lane", "Avenue"}
	postcodes := []string{"SW", "NW", "SE", "E", "W", "N"}
	return NewGenerator(firstnames, lastnames, loremWords, cities, counties, countries, streets, streetTypes, postcodes)
}

func seededRNG(seed int64) RNG {
	return decide.NewSeededRNG(seed)
}

func TestName(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	name, err := g.Name(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	parts := strings.Split(name, " ")
	if len(parts) != 2 {
		t.Errorf("expected name to have two parts, got %q", name)
	}
}

func TestName_Deterministic(t *testing.T) {
	g := testGenerator()

	// Same seed should produce same result
	rng1 := seededRNG(42)
	name1, _ := g.Name(rng1)

	rng2 := seededRNG(42)
	name2, _ := g.Name(rng2)

	if name1 != name2 {
		t.Errorf("expected deterministic output, got %q and %q", name1, name2)
	}
}

func TestEmail_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	email, err := g.Email(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Email should match basic format
	emailRegex := regexp.MustCompile(`^[a-z]+\.[a-z]+@[a-z]+\.[a-z]+$`)
	if !emailRegex.MatchString(email) {
		t.Errorf("email %q does not match expected format", email)
	}
}

func TestPhone_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	phone, err := g.Phone(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// UK mobile phone format: 07XXX XXXXXX
	phoneRegex := regexp.MustCompile(`^07\d{3} \d{6}$`)
	if !phoneRegex.MatchString(phone) {
		t.Errorf("phone %q does not match UK mobile format", phone)
	}
}

func TestAddress_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	addr, err := g.Address(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Address should have number + street name + type
	parts := strings.Split(addr, " ")
	if len(parts) < 3 {
		t.Errorf("address %q should have at least 3 parts", addr)
	}
}

func TestPostcode_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	postcode, err := g.Postcode(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// UK postcode format: outward code (2-4 chars) + space + inward code (digit + 2 letters)
	postcodeRegex := regexp.MustCompile(`^[A-Z]{1,2}\d[A-Z\d]? \d[A-Z]{2}$`)
	if !postcodeRegex.MatchString(postcode) {
		t.Errorf("postcode %q does not match UK format", postcode)
	}
}

func TestDate_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	date, err := g.Date(rng, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	dateRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}$`)
	if !dateRegex.MatchString(date) {
		t.Errorf("date %q does not match YYYY-MM-DD format", date)
	}
}

func TestDatetime_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	dt, err := g.Datetime(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// RFC3339 format
	dtRegex := regexp.MustCompile(`^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}Z$`)
	if !dtRegex.MatchString(dt) {
		t.Errorf("datetime %q does not match RFC3339 format", dt)
	}
}

func TestTime_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	tm, err := g.Time(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	timeRegex := regexp.MustCompile(`^\d{2}:\d{2}:\d{2}$`)
	if !timeRegex.MatchString(tm) {
		t.Errorf("time %q does not match HH:MM:SS format", tm)
	}
}

func TestUUID_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	uuid, err := g.UUID(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// UUID v4 format
	uuidRegex := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-4[0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}$`)
	if !uuidRegex.MatchString(uuid) {
		t.Errorf("uuid %q does not match UUID v4 format", uuid)
	}
}

func TestHex_Length(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	tests := []int{8, 16, 32, 64}
	for _, length := range tests {
		hex, err := g.Hex(rng, length)
		if err != nil {
			t.Fatalf("unexpected error for length %d: %v", length, err)
		}
		if len(hex) != length {
			t.Errorf("expected hex of length %d, got %d", length, len(hex))
		}
	}
}

func TestNumber_Range(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	min, max := 10, 20
	for i := 0; i < 100; i++ {
		n, err := g.Number(rng, min, max)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if n < min || n > max {
			t.Errorf("number %d out of range [%d, %d]", n, min, max)
		}
	}
}

func TestNumber_MinGreaterThanMax(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	_, err := g.Number(rng, 20, 10)
	if err == nil {
		t.Error("expected error when min > max")
	}
}

func TestIPv4_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	ip, err := g.IPv4(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ipRegex := regexp.MustCompile(`^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$`)
	if !ipRegex.MatchString(ip) {
		t.Errorf("ipv4 %q does not match expected format", ip)
	}
}

func TestIPv6_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	ip, err := g.IPv6(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ipRegex := regexp.MustCompile(`^[0-9a-f]{4}(:[0-9a-f]{4}){7}$`)
	if !ipRegex.MatchString(ip) {
		t.Errorf("ipv6 %q does not match expected format", ip)
	}
}

func TestMAC_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	mac, err := g.MAC(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	macRegex := regexp.MustCompile(`^[0-9a-f]{2}(:[0-9a-f]{2}){5}$`)
	if !macRegex.MatchString(mac) {
		t.Errorf("mac %q does not match expected format", mac)
	}
}

func TestURL_Format(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	url, err := g.URL(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	urlRegex := regexp.MustCompile(`^https?://\w+\.\w+\.\w+$`)
	if !urlRegex.MatchString(url) {
		t.Errorf("url %q does not match expected format", url)
	}
}

func TestSentence_Structure(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	sentence, err := g.Sentence(rng)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with capital letter
	if len(sentence) == 0 || sentence[0] < 'A' || sentence[0] > 'Z' {
		t.Errorf("sentence should start with capital letter: %q", sentence)
	}

	// Should end with period
	if !strings.HasSuffix(sentence, ".") {
		t.Errorf("sentence should end with period: %q", sentence)
	}

	// Should have 5-15 words
	words := strings.Split(strings.TrimSuffix(sentence, "."), " ")
	if len(words) < 5 || len(words) > 15 {
		t.Errorf("sentence should have 5-15 words, got %d: %q", len(words), sentence)
	}
}

func TestLorem_DefaultParagraph(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	lorem, err := g.Lorem(rng, 0, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should start with "Lorem ipsum"
	if !strings.HasPrefix(lorem, "Lorem ipsum") {
		t.Errorf("lorem should start with 'Lorem ipsum': %q", lorem[:min(50, len(lorem))])
	}
}

func TestLorem_Words(t *testing.T) {
	g := testGenerator()
	rng := seededRNG(42)

	lorem, err := g.Lorem(rng, 10, 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	words := strings.Split(lorem, " ")
	if len(words) != 10 {
		t.Errorf("expected 10 words, got %d", len(words))
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
