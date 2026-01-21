// Package crypto provides wiki-wide encryption/decryption functionality.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	// File format constants
	magicHeader   = "REGIMENENC" // 10 bytes
	formatVersion = 1            // 1 byte
	nonceSize     = 12           // GCM standard nonce size
	keySize       = 32           // AES-256 key size
	saltSize      = 32           // Salt for Argon2

	// Argon2 parameters (balanced for security and usability)
	argonTime    = 3          // iterations
	argonMemory  = 128 * 1024 // 128 MiB
	argonThreads = 4          // parallelism
)

// Marker represents the .encrypted marker file metadata.
type Marker struct {
	Salt           string `json:"salt"`            // hex-encoded salt
	Argon2Time     uint32 `json:"argon2_time"`     // time parameter
	Argon2Memory   uint32 `json:"argon2_memory"`   // memory in KiB
	Argon2Threads  uint8  `json:"argon2_threads"`  // parallelism
	EncryptedFiles int    `json:"encrypted_files"` // count of encrypted files
}

// EncryptReport contains results of encryption operation.
type EncryptReport struct {
	Encrypted []string // successfully encrypted files
	Failed    []string // files that failed to encrypt
	Skipped   []string // files that were skipped
}

// DecryptReport contains results of decryption operation.
type DecryptReport struct {
	Decrypted []string // successfully decrypted files
	Failed    []string // files that failed to decrypt
	Skipped   []string // files that were skipped
}

// EncryptWiki encrypts all eligible files in the wiki directory.
// Eligible files: .md and .json files (excluding .git/ directory and symlinks).
// Creates .encrypted marker file with encryption metadata.
func EncryptWiki(wikiDir string, passphrase string) (*EncryptReport, error) {
	// Check if already encrypted
	markerPath := filepath.Join(wikiDir, ".encrypted")
	if _, err := os.Stat(markerPath); err == nil {
		return nil, fmt.Errorf("wiki is already encrypted (found %s)", markerPath)
	}

	// Generate salt
	salt := make([]byte, saltSize)
	if _, err := rand.Read(salt); err != nil {
		return nil, fmt.Errorf("failed to generate salt: %w", err)
	}

	// Derive key from passphrase
	key := argon2.IDKey([]byte(passphrase), salt, argonTime, argonMemory, argonThreads, keySize)

	report := &EncryptReport{
		Encrypted: []string{},
		Failed:    []string{},
		Skipped:   []string{},
	}

	// Walk directory tree and encrypt eligible files
	err := filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip .git directory entirely
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			report.Skipped = append(report.Skipped, path)
			return nil
		}

		// Check if file is eligible (.md or .json)
		ext := filepath.Ext(path)
		if ext != ".md" && ext != ".json" {
			return nil
		}

		// Get relative path for AAD (additional authenticated data)
		relPath, err := filepath.Rel(wikiDir, path)
		if err != nil {
			report.Failed = append(report.Failed, path)
			return nil // Continue processing other files
		}

		// Encrypt the file
		if err := encryptFile(path, key, relPath); err != nil {
			report.Failed = append(report.Failed, path)
			return nil // Continue processing other files
		}

		report.Encrypted = append(report.Encrypted, path)
		return nil
	})

	if err != nil {
		return report, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Create marker file
	marker := Marker{
		Salt:           hex.EncodeToString(salt),
		Argon2Time:     argonTime,
		Argon2Memory:   argonMemory,
		Argon2Threads:  argonThreads,
		EncryptedFiles: len(report.Encrypted),
	}

	markerData, err := json.MarshalIndent(marker, "", "  ")
	if err != nil {
		return report, fmt.Errorf("failed to marshal marker: %w", err)
	}

	if err := os.WriteFile(markerPath, markerData, 0644); err != nil {
		return report, fmt.Errorf("failed to write marker file: %w", err)
	}

	return report, nil
}

// DecryptWiki decrypts all encrypted files in the wiki directory.
// Reads .encrypted marker file for decryption metadata.
func DecryptWiki(wikiDir string, passphrase string) (*DecryptReport, error) {
	// Read marker file
	markerPath := filepath.Join(wikiDir, ".encrypted")
	markerData, err := os.ReadFile(markerPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("wiki is not encrypted (no .encrypted marker found)")
		}
		return nil, fmt.Errorf("failed to read marker file: %w", err)
	}

	var marker Marker
	if err := json.Unmarshal(markerData, &marker); err != nil {
		return nil, fmt.Errorf("failed to parse marker file: %w", err)
	}

	// Decode salt
	salt, err := hex.DecodeString(marker.Salt)
	if err != nil {
		return nil, fmt.Errorf("invalid salt in marker file: %w", err)
	}

	// Derive key from passphrase
	key := argon2.IDKey(
		[]byte(passphrase),
		salt,
		marker.Argon2Time,
		marker.Argon2Memory,
		marker.Argon2Threads,
		keySize,
	)

	report := &DecryptReport{
		Decrypted: []string{},
		Failed:    []string{},
		Skipped:   []string{},
	}

	// Walk directory tree and decrypt .enc files
	err = filepath.Walk(wikiDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories
		if info.IsDir() {
			// Skip .git directory entirely
			if info.Name() == ".git" {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip symlinks
		if info.Mode()&os.ModeSymlink != 0 {
			report.Skipped = append(report.Skipped, path)
			return nil
		}

		// Only process .enc files
		if filepath.Ext(path) != ".enc" {
			return nil
		}

		// Get relative path for AAD (strip .enc extension)
		originalPath := strings.TrimSuffix(path, ".enc")
		relPath, err := filepath.Rel(wikiDir, originalPath)
		if err != nil {
			report.Failed = append(report.Failed, path)
			return nil // Continue processing other files
		}

		// Decrypt the file
		if err := decryptFile(path, originalPath, key, relPath); err != nil {
			report.Failed = append(report.Failed, path)
			return nil // Continue processing other files
		}

		report.Decrypted = append(report.Decrypted, path)
		return nil
	})

	if err != nil {
		return report, fmt.Errorf("failed to walk directory: %w", err)
	}

	// Remove marker file if decryption was successful
	if len(report.Failed) == 0 {
		if err := os.Remove(markerPath); err != nil {
			return report, fmt.Errorf("failed to remove marker file: %w", err)
		}
	} else {
		return report, fmt.Errorf("decryption incomplete: %d files failed", len(report.Failed))
	}

	return report, nil
}

// encryptFile encrypts a file in place, replacing it with .enc version.
// File format: REGIMENENC (10 bytes) + version (1 byte) + nonce (12 bytes) + ciphertext
func encryptFile(path string, key []byte, aad string) error {
	// Read plaintext
	plaintext, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Generate nonce
	nonce := make([]byte, nonceSize)
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}

	// Encrypt with AAD (prevents file swapping)
	ciphertext := gcm.Seal(nil, nonce, plaintext, []byte(aad))

	// Build encrypted file content
	encData := make([]byte, 0, len(magicHeader)+1+nonceSize+len(ciphertext))
	encData = append(encData, []byte(magicHeader)...)
	encData = append(encData, formatVersion)
	encData = append(encData, nonce...)
	encData = append(encData, ciphertext...)

	// Write encrypted file
	encPath := path + ".enc"
	if err := os.WriteFile(encPath, encData, 0644); err != nil {
		return fmt.Errorf("failed to write encrypted file: %w", err)
	}

	// Remove original file
	if err := os.Remove(path); err != nil {
		// Try to clean up encrypted file
		os.Remove(encPath)
		return fmt.Errorf("failed to remove original file: %w", err)
	}

	return nil
}

// decryptFile decrypts an encrypted file and restores the original.
func decryptFile(encPath string, origPath string, key []byte, aad string) error {
	// Read encrypted file
	encData, err := os.ReadFile(encPath)
	if err != nil {
		return fmt.Errorf("failed to read encrypted file: %w", err)
	}

	// Validate minimum size
	minSize := len(magicHeader) + 1 + nonceSize
	if len(encData) < minSize {
		return fmt.Errorf("encrypted file too short")
	}

	// Verify magic header
	if subtle.ConstantTimeCompare(encData[:len(magicHeader)], []byte(magicHeader)) != 1 {
		return fmt.Errorf("invalid encrypted file format")
	}

	// Check version
	version := encData[len(magicHeader)]
	if version != formatVersion {
		return fmt.Errorf("unsupported format version: %d", version)
	}

	// Extract nonce and ciphertext
	offset := len(magicHeader) + 1
	nonce := encData[offset : offset+nonceSize]
	ciphertext := encData[offset+nonceSize:]

	// Create cipher
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("failed to create GCM: %w", err)
	}

	// Decrypt with AAD verification
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(aad))
	if err != nil {
		return fmt.Errorf("decryption failed (wrong passphrase or corrupted file): %w", err)
	}

	// Write decrypted file
	if err := os.WriteFile(origPath, plaintext, 0644); err != nil {
		return fmt.Errorf("failed to write decrypted file: %w", err)
	}

	// Remove encrypted file
	if err := os.Remove(encPath); err != nil {
		// Try to clean up decrypted file
		os.Remove(origPath)
		return fmt.Errorf("failed to remove encrypted file: %w", err)
	}

	return nil
}
