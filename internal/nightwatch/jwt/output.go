package jwt

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// OutputOption represents output formatting options.
type OutputOption struct {
	JSON bool
}

// FormatDecode formats the decoded token output.
func FormatDecode(decoded *DecodedToken, opt OutputOption) string {
	if opt.JSON {
		return toJSON(map[string]interface{}{
			"header":    decoded.Header,
			"payload":   decoded.Payload,
			"signature": decoded.Signature,
			"verified":  decoded.Verified,
			"raw":       decoded.Raw,
		})
	}

	var sb strings.Builder
	sb.WriteString("Header:\n")
	formatMap(decoded.Header, &sb, "  ")
	sb.WriteString("\nPayload:\n")
	formatMap(decoded.Payload, &sb, "  ")
	sb.WriteString(fmt.Sprintf("\nSignature: %s\n", decoded.Signature))
	sb.WriteString(fmt.Sprintf("Verified: %v\n", decoded.Verified))
	sb.WriteString("\n\u26a0\ufe0f  WARNING: This token was NOT verified. Use 'verify' command for signature validation.\n")
	return sb.String()
}

// FormatHeader formats header as raw JSON.
func FormatHeader(header map[string]interface{}) string {
	return toJSON(header)
}

// FormatPayload formats payload as raw JSON.
func FormatPayload(payload map[string]interface{}) string {
	return toJSON(payload)
}

// FormatVerify formats verification output.
func FormatVerify(valid bool, algorithm string, header, payload map[string]interface{}, errStr string, opt OutputOption) string {
	if opt.JSON {
		result := map[string]interface{}{
			"valid":     valid,
			"header":    header,
			"payload":   payload,
			"verified":  valid,
			"algorithm": algorithm,
		}
		if !valid && errStr != "" {
			result["error"] = errStr
		}
		return toJSON(result)
	}

	if valid {
		return fmt.Sprintf("\u2713 Token is valid\nAlgorithm: %s\n", algorithm)
	}
	return fmt.Sprintf("\u2717 Token verification failed: %s\n", errStr)
}

// FormatCreate formats token creation output.
func FormatCreate(token string, header, payload map[string]interface{}, opt OutputOption) string {
	if opt.JSON {
		return toJSON(map[string]interface{}{
			"token":   token,
			"header":  header,
			"payload": payload,
		})
	}
	return token + "\n"
}

// FormatExp formats expiration check output.
func FormatExp(result *ExpResult, opt OutputOption) string {
	if opt.JSON {
		out := map[string]interface{}{
			"now":    result.Now.Format(time.RFC3339),
			"status": result.Status,
		}
		if result.IssuedAt != nil {
			out["iat"] = result.IssuedAt.Format(time.RFC3339)
		}
		if result.ExpiresAt != nil {
			out["exp"] = result.ExpiresAt.Format(time.RFC3339)
		}
		if result.Remaining != "" {
			out["remaining"] = result.Remaining
		}
		return toJSON(out)
	}

	var sb strings.Builder
	if result.IssuedAt != nil {
		sb.WriteString(fmt.Sprintf("Issued At:  %s\n", result.IssuedAt.Format(time.RFC3339)))
	}
	if result.ExpiresAt != nil {
		sb.WriteString(fmt.Sprintf("Expires At: %s\n", result.ExpiresAt.Format(time.RFC3339)))
	}
	sb.WriteString(fmt.Sprintf("Now:        %s\n", result.Now.Format(time.RFC3339)))
	sb.WriteString(fmt.Sprintf("Status:     %s", result.Status))
	if result.Remaining != "" {
		sb.WriteString(fmt.Sprintf(" (%s remaining)", result.Remaining))
	}
	sb.WriteString("\n")
	return sb.String()
}

// toJSON marshals a value to indented JSON with a trailing newline.
func toJSON(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b) + "\n"
}

// formatMap formats a map for human-readable output.
func formatMap(m map[string]interface{}, sb *strings.Builder, indent string) {
	for k, v := range m {
		switch val := v.(type) {
		case map[string]interface{}:
			sb.WriteString(fmt.Sprintf("%s%s:\n", indent, k))
			formatMap(val, sb, indent+"  ")
		case float64:
			if val == float64(int64(val)) {
				sb.WriteString(fmt.Sprintf("%s%s: %d\n", indent, k, int64(val)))
			} else {
				sb.WriteString(fmt.Sprintf("%s%s: %g\n", indent, k, val))
			}
		default:
			sb.WriteString(fmt.Sprintf("%s%s: %v\n", indent, k, v))
		}
	}
}
