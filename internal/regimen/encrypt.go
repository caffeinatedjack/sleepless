package regimen

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"syscall"

	"github.com/spf13/cobra"
	"gitlab.com/caffeinatedjack/sleepless/pkg/crypto"
	"golang.org/x/term"
)

var encryptCmd = &cobra.Command{
	Use:   "encrypt",
	Short: "Encrypt all wiki files",
	Long: `Encrypts all .md and .json files in the wiki directory using AES-256-GCM.

The wiki must not already be encrypted. After encryption, all regimen commands
will refuse to run until you decrypt the wiki.

Files are encrypted in place with .enc extension. A .encrypted marker file is
created with encryption metadata.`,
	Example: `  # Encrypt with interactive passphrase prompt
  regimen encrypt

  # Encrypt with passphrase from stdin (for scripting)
  echo "my-passphrase" | regimen encrypt --passphrase-stdin`,
	RunE: runEncrypt,
}

var (
	encryptPassphraseStdin bool
)

func init() {
	encryptCmd.Flags().BoolVar(&encryptPassphraseStdin, "passphrase-stdin", false, "Read passphrase from stdin")
	rootCmd.AddCommand(encryptCmd)
}

func runEncrypt(cmd *cobra.Command, args []string) error {
	wikiDir := getWikiDir()

	// Read passphrase
	var passphrase string
	var err error
	if encryptPassphraseStdin {
		passphrase, err = readPassphraseStdin()
		if err != nil {
			return err
		}
	} else {
		passphrase, err = readPassphraseInteractive(true)
		if err != nil {
			return err
		}
	}

	if passphrase == "" {
		return fmt.Errorf("passphrase cannot be empty")
	}

	fmt.Fprintf(os.Stderr, "Encrypting wiki at %s...\n", wikiDir)

	// Encrypt wiki
	report, err := crypto.EncryptWiki(wikiDir, passphrase)
	if err != nil {
		return fmt.Errorf("encryption failed: %w", err)
	}

	// Print report
	fmt.Fprintf(os.Stderr, "\nEncryption complete:\n")
	fmt.Fprintf(os.Stderr, "  Encrypted: %d files\n", len(report.Encrypted))
	if len(report.Skipped) > 0 {
		fmt.Fprintf(os.Stderr, "  Skipped:   %d files\n", len(report.Skipped))
	}
	if len(report.Failed) > 0 {
		fmt.Fprintf(os.Stderr, "  Failed:    %d files\n", len(report.Failed))
		fmt.Fprintf(os.Stderr, "\nFailed files:\n")
		for _, path := range report.Failed {
			fmt.Fprintf(os.Stderr, "  - %s\n", path)
		}
	}

	fmt.Fprintf(os.Stderr, "\nWiki is now encrypted. Use 'regimen decrypt' to decrypt.\n")

	return nil
}

// readPassphraseStdin reads a passphrase from stdin.
func readPassphraseStdin() (string, error) {
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return "", fmt.Errorf("failed to read passphrase: %w", err)
		}
		return "", fmt.Errorf("no passphrase provided")
	}
	return strings.TrimSpace(scanner.Text()), nil
}

// readPassphraseInteractive reads a passphrase interactively (with confirmation if confirm=true).
func readPassphraseInteractive(confirm bool) (string, error) {
	fmt.Fprint(os.Stderr, "Enter passphrase: ")
	passphrase, err := readPasswordFromTerminal()
	if err != nil {
		return "", fmt.Errorf("failed to read passphrase: %w", err)
	}
	fmt.Fprintln(os.Stderr)

	if confirm {
		fmt.Fprint(os.Stderr, "Confirm passphrase: ")
		confirmation, err := readPasswordFromTerminal()
		if err != nil {
			return "", fmt.Errorf("failed to read confirmation: %w", err)
		}
		fmt.Fprintln(os.Stderr)

		if passphrase != confirmation {
			return "", fmt.Errorf("passphrases do not match")
		}
	}

	return passphrase, nil
}

// readPasswordFromTerminal reads a password from the terminal without echoing.
func readPasswordFromTerminal() (string, error) {
	fd := int(syscall.Stdin)
	if !term.IsTerminal(fd) {
		return "", fmt.Errorf("stdin is not a terminal")
	}

	bytePassword, err := term.ReadPassword(fd)
	if err != nil {
		return "", err
	}

	return string(bytePassword), nil
}
