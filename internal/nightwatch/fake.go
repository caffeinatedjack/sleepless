package nightwatch

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/internal/nightwatch/fake"
)

//go:embed data/lorem.txt
var loremData string

//go:embed data/cities.txt
var citiesData string

//go:embed data/counties.txt
var countiesData string

//go:embed data/countries.txt
var countriesData string

//go:embed data/streets.txt
var streetsData string

//go:embed data/street_types.txt
var streetTypesData string

//go:embed data/postcodes.txt
var postcodesData string

var (
	fakeCount     int
	fakeSeed      *int64
	fakeJSON      bool
	fakeSeparator string
)

// Cached generator
var fakeGenerator *fake.Generator

func init() {
	rootCmd.AddCommand(fakeCmd)

	// Global flags for all fake subcommands
	fakeCmd.PersistentFlags().IntVar(&fakeCount, "count", 1, "Number of items to generate")
	fakeCmd.PersistentFlags().BoolVar(&fakeJSON, "json", false, "Output in JSON format")
	fakeCmd.PersistentFlags().StringVar(&fakeSeparator, "separator", "\n", "Separator for multi-item output")
	fakeCmd.PersistentFlags().Int64("seed", 0, "Random seed for reproducibility")
	fakeCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("seed") {
			seed, _ := cmd.Flags().GetInt64("seed")
			fakeSeed = &seed
		} else {
			fakeSeed = nil
		}
	}

	// Add all subcommands
	fakeCmd.AddCommand(fakeNameCmd)
	fakeCmd.AddCommand(fakeFirstnameCmd)
	fakeCmd.AddCommand(fakeLastnameCmd)
	fakeCmd.AddCommand(fakeEmailCmd)
	fakeCmd.AddCommand(fakeUsernameCmd)
	fakeCmd.AddCommand(fakePhoneCmd)
	fakeCmd.AddCommand(fakeAddressCmd)
	fakeCmd.AddCommand(fakeCityCmd)
	fakeCmd.AddCommand(fakeCountyCmd)
	fakeCmd.AddCommand(fakeCountryCmd)
	fakeCmd.AddCommand(fakePostcodeCmd)
	fakeCmd.AddCommand(fakeDateCmd)
	fakeCmd.AddCommand(fakeDatetimeCmd)
	fakeCmd.AddCommand(fakeTimeCmd)
	fakeCmd.AddCommand(fakeUUIDCmd)
	fakeCmd.AddCommand(fakeHexCmd)
	fakeCmd.AddCommand(fakeNumberCmd)
	fakeCmd.AddCommand(fakeURLCmd)
	fakeCmd.AddCommand(fakeIPv4Cmd)
	fakeCmd.AddCommand(fakeIPv6Cmd)
	fakeCmd.AddCommand(fakeMACCmd)
	fakeCmd.AddCommand(fakeWordCmd)
	fakeCmd.AddCommand(fakeSentenceCmd)
	fakeCmd.AddCommand(fakeParagraphCmd)
	fakeCmd.AddCommand(fakeLoremCmd)
	fakeCmd.AddCommand(fakeTemplateCmd)

	// Date-specific flags
	fakeDateCmd.Flags().Int("past", 0, "Maximum days in the past")
	fakeDateCmd.Flags().Int("future", 0, "Maximum days in the future")

	// Number-specific flags
	fakeNumberCmd.Flags().Int("min", 0, "Minimum value")
	fakeNumberCmd.Flags().Int("max", 100, "Maximum value")

	// Hex-specific flags
	fakeHexCmd.Flags().Int("length", 16, "Length of hex string")

	// Lorem-specific flags
	fakeLoremCmd.Flags().Int("words", 0, "Number of words to generate")
	fakeLoremCmd.Flags().Int("sentences", 0, "Number of sentences to generate")
	fakeLoremCmd.Flags().Int("paragraphs", 0, "Number of paragraphs to generate")
}

// getGenerator returns the cached generator, initializing it if needed.
func getGenerator() (*fake.Generator, error) {
	if fakeGenerator != nil {
		return fakeGenerator, nil
	}

	// Parse all embedded data files
	firstnames := parseNameData(firstnamesData)
	lastnames := parseNameData(surnamesData)
	loremWords := parseNameData(loremData)
	cities := parseNameData(citiesData)
	counties := parseNameData(countiesData)
	countries := parseNameData(countriesData)
	streets := parseNameData(streetsData)
	streetTypes := parseNameData(streetTypesData)
	postcodes := parseNameData(postcodesData)

	fakeGenerator = fake.NewGenerator(firstnames, lastnames, loremWords, cities, counties, countries, streets, streetTypes, postcodes)
	return fakeGenerator, nil
}

// parseNameData parses embedded name data into a slice.
func parseNameData(data string) []string {
	var result []string
	for _, line := range strings.Split(data, "\n") {
		name := strings.TrimSpace(line)
		if name != "" && !strings.HasPrefix(name, "#") {
			result = append(result, name)
		}
	}
	return result
}

// outputResults outputs the results in text or JSON format.
func outputFakeResults(typeName string, results []string) {
	if fakeJSON {
		output := fake.Output{
			Type:    typeName,
			Seed:    fakeSeed,
			Count:   fakeCount,
			Results: results,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(strings.Join(results, fakeSeparator))
		if fakeSeparator == "\n" {
			fmt.Println()
		}
	}
}

// outputTemplateResults outputs template results in text or JSON format.
func outputTemplateResults(template string, results []string) {
	if fakeJSON {
		output := struct {
			Type     string   `json:"type"`
			Template string   `json:"template"`
			Seed     *int64   `json:"seed"`
			Count    int      `json:"count"`
			Results  []string `json:"results"`
		}{
			Type:     "template",
			Template: template,
			Seed:     fakeSeed,
			Count:    fakeCount,
			Results:  results,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		fmt.Print(strings.Join(results, fakeSeparator))
		if fakeSeparator == "\n" {
			fmt.Println()
		}
	}
}

var fakeCmd = &cobra.Command{
	Use:   "fake",
	Short: "Generate fake data for testing and development",
	Long: `Generate fake/placeholder data for development, testing, and demonstration.

Supports various data types including names, emails, addresses, dates, and more.
Use --seed for reproducible output and --json for machine-readable format.

Examples:
    nightwatch fake name
    nightwatch fake email --count 5
    nightwatch fake number --min 1 --max 100
    nightwatch fake template "{{name}} <{{email}}>"
    nightwatch fake name --count 3 --seed 42`,
}

var fakeNameCmd = &cobra.Command{
	Use:   "name",
	Short: "Generate a random full name",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Name(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("name", results)
		return nil
	},
}

var fakeFirstnameCmd = &cobra.Command{
	Use:   "firstname",
	Short: "Generate a random first name",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Firstname(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("firstname", results)
		return nil
	},
}

var fakeLastnameCmd = &cobra.Command{
	Use:   "lastname",
	Short: "Generate a random last name",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Lastname(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("lastname", results)
		return nil
	},
}

var fakeEmailCmd = &cobra.Command{
	Use:   "email",
	Short: "Generate a random email address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Email(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("email", results)
		return nil
	},
}

var fakeUsernameCmd = &cobra.Command{
	Use:   "username",
	Short: "Generate a random username",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Username(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("username", results)
		return nil
	},
}

var fakePhoneCmd = &cobra.Command{
	Use:   "phone",
	Short: "Generate a random phone number",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Phone(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("phone", results)
		return nil
	},
}

var fakeAddressCmd = &cobra.Command{
	Use:   "address",
	Short: "Generate a random street address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Address(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("address", results)
		return nil
	},
}

var fakeCityCmd = &cobra.Command{
	Use:   "city",
	Short: "Generate a random city name",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.City(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("city", results)
		return nil
	},
}

var fakeCountyCmd = &cobra.Command{
	Use:   "county",
	Short: "Generate a random UK county",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.County(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("county", results)
		return nil
	},
}

var fakeCountryCmd = &cobra.Command{
	Use:   "country",
	Short: "Generate a random country name",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Country(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("country", results)
		return nil
	},
}

var fakePostcodeCmd = &cobra.Command{
	Use:   "postcode",
	Short: "Generate a random UK postcode",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Postcode(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("postcode", results)
		return nil
	},
}

var fakeDateCmd = &cobra.Command{
	Use:   "date",
	Short: "Generate a random date (YYYY-MM-DD)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}

		past, _ := cmd.Flags().GetInt("past")
		future, _ := cmd.Flags().GetInt("future")

		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Date(rng, past, future)
			if err != nil {
				return err
			}
		}
		outputFakeResults("date", results)
		return nil
	},
}

var fakeDatetimeCmd = &cobra.Command{
	Use:   "datetime",
	Short: "Generate a random RFC3339 timestamp",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Datetime(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("datetime", results)
		return nil
	},
}

var fakeTimeCmd = &cobra.Command{
	Use:   "time",
	Short: "Generate a random time (HH:MM:SS)",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Time(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("time", results)
		return nil
	},
}

var fakeUUIDCmd = &cobra.Command{
	Use:   "uuid",
	Short: "Generate a random UUID v4",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.UUID(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("uuid", results)
		return nil
	},
}

var fakeHexCmd = &cobra.Command{
	Use:   "hex",
	Short: "Generate a random hexadecimal string",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}

		length, _ := cmd.Flags().GetInt("length")

		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Hex(rng, length)
			if err != nil {
				return err
			}
		}
		outputFakeResults("hex", results)
		return nil
	},
}

var fakeNumberCmd = &cobra.Command{
	Use:   "number",
	Short: "Generate a random number",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}

		min, _ := cmd.Flags().GetInt("min")
		max, _ := cmd.Flags().GetInt("max")

		if min > max {
			return fmt.Errorf("min cannot be greater than max")
		}

		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			n, err := g.Number(rng, min, max)
			if err != nil {
				return err
			}
			results[i] = strconv.Itoa(n)
		}
		outputFakeResults("number", results)
		return nil
	},
}

var fakeURLCmd = &cobra.Command{
	Use:   "url",
	Short: "Generate a random URL",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.URL(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("url", results)
		return nil
	},
}

var fakeIPv4Cmd = &cobra.Command{
	Use:   "ipv4",
	Short: "Generate a random IPv4 address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.IPv4(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("ipv4", results)
		return nil
	},
}

var fakeIPv6Cmd = &cobra.Command{
	Use:   "ipv6",
	Short: "Generate a random IPv6 address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.IPv6(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("ipv6", results)
		return nil
	},
}

var fakeMACCmd = &cobra.Command{
	Use:   "mac",
	Short: "Generate a random MAC address",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.MAC(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("mac", results)
		return nil
	},
}

var fakeWordCmd = &cobra.Command{
	Use:   "word",
	Short: "Generate a random word",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Word(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("word", results)
		return nil
	},
}

var fakeSentenceCmd = &cobra.Command{
	Use:   "sentence",
	Short: "Generate a random sentence",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Sentence(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("sentence", results)
		return nil
	},
}

var fakeParagraphCmd = &cobra.Command{
	Use:   "paragraph",
	Short: "Generate a random paragraph",
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}
		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Paragraph(rng)
			if err != nil {
				return err
			}
		}
		outputFakeResults("paragraph", results)
		return nil
	},
}

var fakeLoremCmd = &cobra.Command{
	Use:   "lorem",
	Short: "Generate lorem ipsum text",
	Long: `Generate lorem ipsum placeholder text.

By default generates 1 paragraph. Use flags to specify:
  --words N       Generate N words
  --sentences N   Generate N sentences  
  --paragraphs N  Generate N paragraphs

Examples:
    nightwatch fake lorem
    nightwatch fake lorem --words 10
    nightwatch fake lorem --sentences 3
    nightwatch fake lorem --paragraphs 2`,
	RunE: func(cmd *cobra.Command, args []string) error {
		g, err := getGenerator()
		if err != nil {
			return err
		}

		words, _ := cmd.Flags().GetInt("words")
		sentences, _ := cmd.Flags().GetInt("sentences")
		paragraphs, _ := cmd.Flags().GetInt("paragraphs")

		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = g.Lorem(rng, words, sentences, paragraphs)
			if err != nil {
				return err
			}
		}
		outputFakeResults("lorem", results)
		return nil
	},
}

var fakeTemplateCmd = &cobra.Command{
	Use:   "template <string>",
	Short: "Generate data from a template",
	Long: `Generate data by replacing template placeholders with fake data.

Placeholders use the format {{type}} or {{type:arg1:arg2}}.

Supported types:
  name, firstname, lastname, email, username, phone
  address, city, county, country, postcode
  date, datetime, time
  uuid, hex, number
  lorem, word, sentence, paragraph
  url, ipv4, ipv6, mac

Parameterized placeholders:
  {{number:min:max}}       - number in range
  {{lorem:words:N}}        - N words
  {{lorem:sentences:N}}    - N sentences
  {{date:past:N}}          - date up to N days ago
  {{date:future:N}}        - date up to N days ahead
  {{hex:N}}                - hex string of length N

Examples:
    nightwatch fake template "{{name}} <{{email}}>"
    nightwatch fake template "{{firstname}} lives in {{city}}, {{county}}"
    nightwatch fake template '{"name":"{{name}}","age":{{number:18:65}}}'`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		templateStr := args[0]

		tmpl, err := fake.ParseTemplate(templateStr)
		if err != nil {
			return fmt.Errorf("failed to parse template: %w", err)
		}

		if err := tmpl.Validate(); err != nil {
			return err
		}

		g, err := getGenerator()
		if err != nil {
			return err
		}

		rng := fake.NewRNG(fakeSeed)
		results := make([]string, fakeCount)
		for i := 0; i < fakeCount; i++ {
			results[i], err = tmpl.Render(g, rng)
			if err != nil {
				return err
			}
		}

		outputTemplateResults(templateStr, results)
		return nil
	},
}
