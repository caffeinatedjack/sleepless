package jwt

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	gojwt "github.com/golang-jwt/jwt/v5"
)

// CreateResult represents token creation output.
type CreateResult struct {
	Token   string                 `json:"token"`
	Header  map[string]interface{} `json:"header"`
	Payload map[string]interface{} `json:"payload"`
}

// CreateToken creates a new JWT with the given parameters.
func CreateToken(alg, secret, keyPath string, claims map[string]interface{}, claimsJSON, iss, sub, aud, expDur, nbfDur, iatStr, jti string) (*CreateResult, error) {
	customClaims, err := buildClaims(claims, claimsJSON, iss, sub, aud, expDur, nbfDur, iatStr, jti)
	if err != nil {
		return nil, err
	}

	method, signingKey, err := getSigningCredentials(alg, secret, keyPath)
	if err != nil {
		return nil, err
	}

	token := gojwt.NewWithClaims(method, customClaims)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		return nil, NewError(ErrKeyLoad, "failed to sign token", err)
	}

	decoded, err := DecodeWithoutVerification(tokenString)
	if err != nil {
		return nil, err
	}

	return &CreateResult{
		Token:   tokenString,
		Header:  decoded.Header,
		Payload: decoded.Payload,
	}, nil
}

// buildClaims constructs the claims map from parameters.
func buildClaims(claims map[string]interface{}, claimsJSON, iss, sub, aud, expDur, nbfDur, iatStr, jti string) (gojwt.MapClaims, error) {
	if claimsJSON != "" {
		var result gojwt.MapClaims
		if err := json.Unmarshal([]byte(claimsJSON), &result); err != nil {
			return nil, NewError(ErrInvalidFormat, "invalid JSON in --payload", err)
		}
		return result, nil
	}

	result := gojwt.MapClaims{}
	for k, v := range claims {
		result[k] = v
	}

	if iss != "" {
		result["iss"] = iss
	}
	if sub != "" {
		result["sub"] = sub
	}
	if aud != "" {
		result["aud"] = aud
	}
	if jti != "" {
		result["jti"] = jti
	}

	now := time.Now()

	if iatStr != "" {
		iatTime, err := parseTimestamp(iatStr)
		if err != nil {
			return nil, NewError(ErrInvalidFormat, "invalid iat timestamp", err)
		}
		result["iat"] = iatTime.Unix()
	} else {
		result["iat"] = now.Unix()
	}

	if nbfDur != "" {
		dur, err := ParseDuration(nbfDur)
		if err != nil {
			return nil, NewError(ErrInvalidFormat, "invalid nbf duration", err)
		}
		result["nbf"] = now.Add(dur).Unix()
	}

	if expDur != "" {
		dur, err := ParseDuration(expDur)
		if err != nil {
			return nil, NewError(ErrInvalidFormat, "invalid exp duration", err)
		}
		result["exp"] = now.Add(dur).Unix()
	}

	return result, nil
}

// getSigningCredentials determines the signing method and key.
func getSigningCredentials(alg, secret, keyPath string) (gojwt.SigningMethod, interface{}, error) {
	if secret != "" {
		return getHMACCredentials(alg, secret)
	}
	if keyPath != "" {
		return getAsymmetricCredentials(alg, keyPath)
	}
	return nil, nil, NewError(ErrKeyLoad, "either --secret or --key is required", nil)
}

func getHMACCredentials(alg, secret string) (gojwt.SigningMethod, interface{}, error) {
	if alg == "" {
		alg = "HS256"
	}
	if GetAlgorithmType(alg) != "HMAC" {
		return nil, nil, NewError(ErrKeyLoad, fmt.Sprintf("algorithm %s requires --key, not --secret", alg), nil)
	}
	return GetSigningMethod(alg), []byte(secret), nil
}

func getAsymmetricCredentials(alg, keyPath string) (gojwt.SigningMethod, interface{}, error) {
	key, err := LoadPrivateKey(keyPath)
	if err != nil {
		return nil, nil, err
	}

	if alg == "" {
		alg = "RS256"
	}

	method := GetSigningMethod(alg)
	if method == nil {
		return nil, nil, NewError(ErrUnsupportedAlgorithm, fmt.Sprintf("unsupported algorithm: %s", alg), nil)
	}

	return method, key, nil
}

// parseTimestamp parses a Unix timestamp or RFC3339 string.
func parseTimestamp(s string) (time.Time, error) {
	if unix, err := strconv.ParseInt(s, 10, 64); err == nil {
		return time.Unix(unix, 0), nil
	}
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t, nil
	}
	return time.Time{}, fmt.Errorf("expected Unix timestamp or RFC3339 format")
}
