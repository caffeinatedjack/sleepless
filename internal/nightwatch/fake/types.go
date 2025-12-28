package fake

// Output represents the JSON output structure for fake data generation.
type Output struct {
	Type     string   `json:"type"`
	Seed     *int64   `json:"seed"`
	Count    int      `json:"count"`
	Template string   `json:"template,omitempty"`
	Results  []string `json:"results"`
}
