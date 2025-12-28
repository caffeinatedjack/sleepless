package regimen

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/decide"
	"gitlab.com/caffeinatedjack/sleepless/pkg/ui"
)

var (
	decideFrom  string
	decideCount int
	decideSeed  *int64
	decideJSON  bool
)

// DecideResult for JSON output
type DecideResult struct {
	Mode    string      `json:"mode"`
	Seed    *int64      `json:"seed"`
	Count   int         `json:"count"`
	Results interface{} `json:"results"`
}

// RollResult for dice roll JSON output
type RollResult struct {
	Notation string `json:"notation"`
	Rolls    []int  `json:"rolls"`
	Total    int    `json:"total"`
}

var decideCmd = &cobra.Command{
	Use:   "decide [options...]",
	Short: "Random choice utilities",
	Long: `Random choice utilities to reduce decision paralysis.

Commands:
    decide <options>           Pick one option at random
    decide pick <n> <options>  Pick n unique options
    decide shuffle <options>   Shuffle options randomly
    decide roll <NdM>          Roll dice (e.g., 2d6)
    decide coin                Flip a coin
    decide number <min> <max>  Random integer in range
    decide weighted <opt:wt>   Weighted random selection

Examples:
    regimen decide pizza tacos sushi
    regimen decide pick 2 alice bob charlie
    regimen decide shuffle a b c d e
    regimen decide roll 2d6
    regimen decide coin
    regimen decide number 1 100
    regimen decide weighted "a:3" "b:1"`,
	Args: cobra.MinimumNArgs(1),
	Run:  runDecidePickOne,
}

func init() {
	rootCmd.AddCommand(decideCmd)

	decideCmd.PersistentFlags().StringVar(&decideFrom, "from", "", "Read options from file")
	decideCmd.PersistentFlags().IntVar(&decideCount, "count", 1, "Repeat operation n times")
	decideCmd.PersistentFlags().BoolVar(&decideJSON, "json", false, "Output as JSON")

	decideCmd.PersistentFlags().Int64("seed", 0, "Random seed for reproducibility")
	decideCmd.PersistentPreRun = func(cmd *cobra.Command, args []string) {
		if cmd.Flags().Changed("seed") {
			seed, _ := cmd.Flags().GetInt64("seed")
			decideSeed = &seed
		} else {
			decideSeed = nil
		}
	}

	decideCmd.AddCommand(pickCmd)
	decideCmd.AddCommand(shuffleCmd)
	decideCmd.AddCommand(rollCmd)
	decideCmd.AddCommand(coinCmd)
	decideCmd.AddCommand(numberCmd)
	decideCmd.AddCommand(weightedCmd)
}

func getOptions(args []string) ([]string, error) {
	if decideFrom != "" && len(args) > 0 {
		return nil, fmt.Errorf("cannot use both positional arguments and --from")
	}

	if decideFrom != "" {
		return readOptionsFromFile(decideFrom)
	}

	if len(args) == 0 {
		return nil, fmt.Errorf("no options provided")
	}

	return args, nil
}

func readOptionsFromFile(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var options []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		options = append(options, line)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}

	if len(options) == 0 {
		return nil, fmt.Errorf("no options found in file")
	}

	return options, nil
}

func outputResult(mode string, results interface{}) {
	if decideJSON {
		output := DecideResult{
			Mode:    mode,
			Seed:    decideSeed,
			Count:   decideCount,
			Results: results,
		}
		data, _ := json.MarshalIndent(output, "", "  ")
		fmt.Println(string(data))
	} else {
		switch v := results.(type) {
		case []string:
			for _, r := range v {
				fmt.Println(r)
			}
		case []int:
			for _, r := range v {
				fmt.Println(r)
			}
		case [][]string:
			for i, run := range v {
				if decideCount > 1 {
					fmt.Printf("Run %d:\n", i+1)
				}
				for _, r := range run {
					if decideCount > 1 {
						fmt.Printf("  %s\n", r)
					} else {
						fmt.Println(r)
					}
				}
			}
		case []RollResult:
			for i, r := range v {
				if decideCount > 1 {
					fmt.Printf("Run %d: ", i+1)
				}
				if len(r.Rolls) == 1 {
					fmt.Printf("%d\n", r.Total)
				} else {
					rollStrs := make([]string, len(r.Rolls))
					for j, roll := range r.Rolls {
						rollStrs[j] = strconv.Itoa(roll)
					}
					fmt.Printf("%s = %d\n", strings.Join(rollStrs, " + "), r.Total)
				}
			}
		default:
			fmt.Printf("%v\n", results)
		}
	}
}

func runDecidePickOne(cmd *cobra.Command, args []string) {
	options, err := getOptions(args)
	if err != nil {
		ui.Error(err.Error())
		os.Exit(1)
	}

	rng := decide.NewRNG(decideSeed)
	results := make([]string, decideCount)

	for i := 0; i < decideCount; i++ {
		idx, err := rng.Intn(len(options))
		if err != nil {
			ui.Error(fmt.Sprintf("Random generation failed: %v", err))
			os.Exit(1)
		}
		results[i] = options[idx]
	}

	outputResult("pickOne", results)
}

var pickCmd = &cobra.Command{
	Use:   "pick <n> [options...]",
	Short: "Pick n unique options at random",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		n, err := strconv.Atoi(args[0])
		if err != nil || n < 1 {
			ui.Error("First argument must be a positive integer")
			os.Exit(1)
		}

		options, err := getOptions(args[1:])
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		if n > len(options) {
			ui.Error(fmt.Sprintf("Cannot pick %d items from %d options", n, len(options)))
			os.Exit(1)
		}

		rng := decide.NewRNG(decideSeed)
		results := make([][]string, decideCount)

		for run := 0; run < decideCount; run++ {
			indices := make([]int, len(options))
			for i := range indices {
				indices[i] = i
			}

			picked := make([]string, n)
			for i := 0; i < n; i++ {
				j, err := rng.Intn(len(indices) - i)
				if err != nil {
					ui.Error(fmt.Sprintf("Random generation failed: %v", err))
					os.Exit(1)
				}
				j += i
				indices[i], indices[j] = indices[j], indices[i]
				picked[i] = options[indices[i]]
			}
			results[run] = picked
		}

		outputResult("pickN", results)
	},
}

var shuffleCmd = &cobra.Command{
	Use:   "shuffle [options...]",
	Short: "Shuffle options randomly",
	Args:  cobra.MinimumNArgs(0),
	Run: func(cmd *cobra.Command, args []string) {
		options, err := getOptions(args)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		rng := decide.NewRNG(decideSeed)
		results := make([][]string, decideCount)

		for run := 0; run < decideCount; run++ {
			shuffled := make([]string, len(options))
			copy(shuffled, options)

			for i := len(shuffled) - 1; i > 0; i-- {
				j, err := rng.Intn(i + 1)
				if err != nil {
					ui.Error(fmt.Sprintf("Random generation failed: %v", err))
					os.Exit(1)
				}
				shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
			}
			results[run] = shuffled
		}

		outputResult("shuffle", results)
	},
}

var diceRegex = regexp.MustCompile(`^(\d+)d(\d+)$`)

var rollCmd = &cobra.Command{
	Use:   "roll <NdM>",
	Short: "Roll dice (e.g., 2d6)",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		notation := strings.ToLower(args[0])
		matches := diceRegex.FindStringSubmatch(notation)
		if matches == nil {
			ui.Error("Invalid dice notation (use NdM, e.g., 2d6)")
			os.Exit(1)
		}

		numDice, _ := strconv.Atoi(matches[1])
		numSides, _ := strconv.Atoi(matches[2])

		if numDice < 1 || numSides < 1 {
			ui.Error("Number of dice and sides must be positive")
			os.Exit(1)
		}

		rng := decide.NewRNG(decideSeed)
		results := make([]RollResult, decideCount)

		for run := 0; run < decideCount; run++ {
			rolls := make([]int, numDice)
			total := 0
			for i := 0; i < numDice; i++ {
				roll, err := rng.Intn(numSides)
				if err != nil {
					ui.Error(fmt.Sprintf("Random generation failed: %v", err))
					os.Exit(1)
				}
				rolls[i] = roll + 1
				total += rolls[i]
			}
			results[run] = RollResult{
				Notation: notation,
				Rolls:    rolls,
				Total:    total,
			}
		}

		outputResult("roll", results)
	},
}

var coinCmd = &cobra.Command{
	Use:   "coin",
	Short: "Flip a coin",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		rng := decide.NewRNG(decideSeed)
		results := make([]string, decideCount)

		for i := 0; i < decideCount; i++ {
			n, err := rng.Intn(2)
			if err != nil {
				ui.Error(fmt.Sprintf("Random generation failed: %v", err))
				os.Exit(1)
			}
			if n == 0 {
				results[i] = "heads"
			} else {
				results[i] = "tails"
			}
		}

		outputResult("coin", results)
	},
}

var numberCmd = &cobra.Command{
	Use:   "number <min> <max>",
	Short: "Random integer in range [min, max]",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		min, err := strconv.Atoi(args[0])
		if err != nil {
			ui.Error("min must be an integer")
			os.Exit(1)
		}

		max, err := strconv.Atoi(args[1])
		if err != nil {
			ui.Error("max must be an integer")
			os.Exit(1)
		}

		if min > max {
			ui.Error("min cannot be greater than max")
			os.Exit(1)
		}

		rng := decide.NewRNG(decideSeed)
		results := make([]int, decideCount)

		for i := 0; i < decideCount; i++ {
			n, err := rng.Intn(max - min + 1)
			if err != nil {
				ui.Error(fmt.Sprintf("Random generation failed: %v", err))
				os.Exit(1)
			}
			results[i] = n + min
		}

		outputResult("number", results)
	},
}

var weightedCmd = &cobra.Command{
	Use:   "weighted <option:weight>...",
	Short: "Weighted random selection",
	Long: `Weighted random selection where probability is proportional to weight.

Examples:
    regimen decide weighted "pizza:3" "tacos:1"
    regimen decide weighted "a:5" "b:3" "c:2"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		options, err := getOptions(args)
		if err != nil {
			ui.Error(err.Error())
			os.Exit(1)
		}

		type weightedOption struct {
			name   string
			weight int
		}

		var weighted []weightedOption
		totalWeight := 0

		for _, opt := range options {
			parts := strings.Split(opt, ":")
			if len(parts) != 2 {
				ui.Error(fmt.Sprintf("Invalid format '%s' (use option:weight)", opt))
				os.Exit(1)
			}

			name := parts[0]
			weight, err := strconv.Atoi(parts[1])
			if err != nil || weight <= 0 {
				ui.Error(fmt.Sprintf("Weight must be a positive integer: %s", parts[1]))
				os.Exit(1)
			}

			weighted = append(weighted, weightedOption{name: name, weight: weight})
			totalWeight += weight
		}

		rng := decide.NewRNG(decideSeed)
		results := make([]string, decideCount)

		for i := 0; i < decideCount; i++ {
			n, err := rng.Intn(totalWeight)
			if err != nil {
				ui.Error(fmt.Sprintf("Random generation failed: %v", err))
				os.Exit(1)
			}

			cumulative := 0
			for _, wo := range weighted {
				cumulative += wo.weight
				if n < cumulative {
					results[i] = wo.name
					break
				}
			}
		}

		outputResult("weighted", results)
	},
}
