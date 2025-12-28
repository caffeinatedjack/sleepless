package jwt

import (
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// LoadPrivateKey loads a PEM-encoded private key for signing.
func LoadPrivateKey(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewError(ErrKeyLoad, fmt.Sprintf("failed to read key file: %s", path), err)
	}

	checkKeyPermissions(path)

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, NewError(ErrKeyLoad, "invalid PEM format", nil)
	}

	// Try PKCS1 RSA private key
	if key, err := x509.ParsePKCS1PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS8 (handles RSA and ECDSA)
	if key, err := x509.ParsePKCS8PrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try EC private key
	if key, err := x509.ParseECPrivateKey(block.Bytes); err == nil {
		return key, nil
	}

	return nil, NewError(ErrKeyLoad, "unable to parse private key (supported formats: PKCS1, PKCS8, EC)", nil)
}

// LoadPublicKey loads a PEM-encoded public key for verification.
func LoadPublicKey(path string) (interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, NewError(ErrKeyLoad, fmt.Sprintf("failed to read key file: %s", path), err)
	}

	block, _ := pem.Decode(data)
	if block == nil {
		return nil, NewError(ErrKeyLoad, "invalid PEM format", nil)
	}

	// Try PKIX public key (most common)
	if key, err := x509.ParsePKIXPublicKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try PKCS1 public key
	if key, err := x509.ParsePKCS1PublicKey(block.Bytes); err == nil {
		return key, nil
	}

	// Try X.509 certificate and extract public key
	if cert, err := x509.ParseCertificate(block.Bytes); err == nil {
		return cert.PublicKey, nil
	}

	return nil, NewError(ErrKeyLoad, "unable to parse public key (supported formats: PKIX, PKCS1, X.509)", nil)
}

// checkKeyPermissions warns if a private key file has insecure permissions.
func checkKeyPermissions(path string) {
	info, err := os.Stat(path)
	if err != nil {
		return
	}
	if info.Mode().Perm()&0077 != 0 {
		fmt.Fprintf(os.Stderr, "Warning: key file %s has insecure permissions (should be 0600)\n", path)
	}
}

// GetSigningMethod returns the jwt.SigningMethod for a given algorithm.
func GetSigningMethod(alg string) gojwt.SigningMethod {
	switch alg {
	case "HS256":
		return gojwt.SigningMethodHS256
	case "HS384":
		return gojwt.SigningMethodHS384
	case "HS512":
		return gojwt.SigningMethodHS512
	case "RS256":
		return gojwt.SigningMethodRS256
	case "RS384":
		return gojwt.SigningMethodRS384
	case "RS512":
		return gojwt.SigningMethodRS512
	case "ES256":
		return gojwt.SigningMethodES256
	case "ES384":
		return gojwt.SigningMethodES384
	case "ES512":
		return gojwt.SigningMethodES512
	default:
		return nil
	}
}
