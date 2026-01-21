package regimen

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/crypto"
)

var decryptCmd = &cobra.Command{
	Use:   "decrypt",
	Short: "Decrypt all wiki files",
	Long: `Decrypts all .enc files in the wiki directory using the passphrase.

The wiki must be encrypted (have a .encrypted marker file). After decryption,
all regimen commands will work normally.

Files are decrypted in place, removing the .enc extension. The .encrypted
marker file is removed after successful decryption.`,
	Example: `  # Decrypt with interactive passphrase prompt
  regimen decrypt

  # Decrypt with passphrase from stdin (for scripting)
  echo "my-passphrase" | regimen decrypt --passphrase-stdin`,
	RunE: runDecrypt,
}

var (
	decryptPassphraseStdin bool
)

func init() {
	decryptCmd.Flags().BoolVar(&decryptPassphraseStdin, "passphrase-stdin", false, "Read passphrase from stdin")
	rootCmd.AddCommand(decryptCmd)
}

func runDecrypt(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()

	// Read passphrase
	var passphrase string
	var err error
	if decryptPassphraseStdin {
		passphrase, err = readPassphraseStdin()
		if err != nil {
			return err
		}
	} else {
		passphrase, err = readPassphraseInteractive(false)
		if err != nil {
			return err
		}
	}

	if passphrase == "" {
		return fmt.Errorf("passphrase cannot be empty")
	}

	fmt.Fprintf(os.Stderr, "Decrypting wiki at %s...\n", wikiDir)

	// Decrypt wiki
	report, err := crypto.DecryptWiki(wikiDir, passphrase)
	if err != nil {
		return fmt.Errorf("decryption failed: %w", err)
	}

	// Print report
	fmt.Fprintf(os.Stderr, "\nDecryption complete:\n")
	fmt.Fprintf(os.Stderr, "  Decrypted: %d files\n", len(report.Decrypted))
	if len(report.Skipped) > 0 {
		fmt.Fprintf(os.Stderr, "  Skipped:   %d files\n", len(report.Skipped))
	}
	if len(report.Failed) > 0 {
		fmt.Fprintf(os.Stderr, "  Failed:    %d files\n", len(report.Failed))
		fmt.Fprintf(os.Stderr, "\nFailed files:\n")
		for _, path := range report.Failed {
			fmt.Fprintf(os.Stderr, "  - %s\n", path)
		}
		return fmt.Errorf("decryption incomplete: %d files failed", len(report.Failed))
	}

	fmt.Fprintf(os.Stderr, "\nWiki is now decrypted. All regimen commands are available.\n")

	return nil
}
