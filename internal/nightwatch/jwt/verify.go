package jwt

import (
	"fmt"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// ExpResult represents expiration check output.
type ExpResult struct {
	IssuedAt  *time.Time
	ExpiresAt *time.Time
	Now       time.Time
	Status    string // "valid", "expired", "no-expiry", "not-yet-valid"
	Remaining string
}

// CheckExpiration checks the expiration status of a token without verification.
func CheckExpiration(token string) (*ExpResult, error) {
	decoded, err := DecodeWithoutVerification(token)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	result := &ExpResult{Now: now}

	if iatVal, ok := decoded.Payload["iat"].(float64); ok {
		iat := time.Unix(int64(iatVal), 0)
		result.IssuedAt = &iat
	}

	if expVal, ok := decoded.Payload["exp"].(float64); ok {
		exp := time.Unix(int64(expVal), 0)
		result.ExpiresAt = &exp

		if now.After(exp) {
			result.Status = "expired"
		} else {
			result.Status = "valid"
			result.Remaining = exp.Sub(now).String()
		}
	} else {
		result.Status = "no-expiry"
	}

	if nbfVal, ok := decoded.Payload["nbf"].(float64); ok {
		nbf := time.Unix(int64(nbfVal), 0)
		if now.Before(nbf) {
			result.Status = "not-yet-valid"
			result.Remaining = nbf.Sub(now).String()
		}
	}

	return result, nil
}

// VerifyToken verifies a JWT signature and claims.
func VerifyToken(token, secret, keyPath string) (valid bool, errMsg string, header, payload map[string]interface{}, err error) {
	decoded, err := DecodeWithoutVerification(token)
	if err != nil {
		return false, "", nil, nil, err
	}

	alg, ok := decoded.Header["alg"].(string)
	if !ok {
		return false, "", nil, nil, NewError(ErrInvalidFormat, "missing algorithm in header", nil)
	}

	if err := ValidateAlgorithm(alg); err != nil {
		return false, "", nil, nil, err
	}

	keyFunc, err := buildKeyFunc(alg, secret, keyPath)
	if err != nil {
		return false, "", nil, nil, err
	}

	parsedToken, err := gojwt.ParseWithClaims(token, &gojwt.MapClaims{}, keyFunc)
	if err != nil {
		return false, "signature verification failed", decoded.Header, decoded.Payload,
			NewError(ErrVerificationFailed, "signature verification failed", err)
	}

	if !parsedToken.Valid {
		return false, "token is invalid", decoded.Header, decoded.Payload,
			NewError(ErrVerificationFailed, "token is invalid", nil)
	}

	claims := parsedToken.Claims.(*gojwt.MapClaims)
	if err := validateTimeClaims(claims); err != nil {
		return false, err.Message, decoded.Header, decoded.Payload, err
	}

	return true, "", decoded.Header, decoded.Payload, nil
}

// buildKeyFunc creates the key function for token verification.
func buildKeyFunc(alg, secret, keyPath string) (gojwt.Keyfunc, error) {
	if secret != "" {
		if GetAlgorithmType(alg) != "HMAC" {
			return nil, NewError(ErrKeyLoad, fmt.Sprintf("algorithm %s requires --key, not --secret", alg), nil)
		}
		return func(t *gojwt.Token) (interface{}, error) {
			if t.Method.Alg() != alg {
				return nil, NewError(ErrVerificationFailed, fmt.Sprintf("algorithm mismatch: expected %s, got %s", alg, t.Method.Alg()), nil)
			}
			return []byte(secret), nil
		}, nil
	}

	if keyPath != "" {
		if GetAlgorithmType(alg) == "HMAC" {
			return nil, NewError(ErrKeyLoad, fmt.Sprintf("algorithm %s requires --secret, not --key", alg), nil)
		}
		publicKey, err := LoadPublicKey(keyPath)
		if err != nil {
			return nil, err
		}
		return func(t *gojwt.Token) (interface{}, error) {
			if t.Method.Alg() != alg {
				return nil, NewError(ErrVerificationFailed, fmt.Sprintf("algorithm mismatch: expected %s, got %s", alg, t.Method.Alg()), nil)
			}
			return publicKey, nil
		}, nil
	}

	return nil, NewError(ErrKeyLoad, "either --secret or --key is required", nil)
}

// validateTimeClaims checks exp and nbf claims.
func validateTimeClaims(claims *gojwt.MapClaims) *JWTError {
	now := time.Now()

	if expVal, ok := (*claims)["exp"].(float64); ok {
		if now.After(time.Unix(int64(expVal), 0)) {
			return NewError(ErrExpired, "token is expired", nil)
		}
	}

	if nbfVal, ok := (*claims)["nbf"].(float64); ok {
		if now.Before(time.Unix(int64(nbfVal), 0)) {
			return NewError(ErrNotYetValid, "token is not yet valid", nil)
		}
	}

	return nil
}
