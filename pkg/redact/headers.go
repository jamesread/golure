package redact

import "strings"

const headerMask = "[redacted]"

// IsSensitiveHeader reports whether an HTTP header named name carries
// credentials and must not be logged. The name is matched case-insensitively.
func IsSensitiveHeader(name string) bool {
	switch strings.ToLower(name) {
	case "authorization", "proxy-authorization", "cookie", "set-cookie", "x-api-key", "x-forwarded-access-token":
		return true
	default:
		return false
	}
}

// HeaderValues returns values safe to log for the header named name. Values of
// sensitive headers are masked, and other values are returned unchanged.
func HeaderValues(name string, values []string) []string {
	if !IsSensitiveHeader(name) {
		return values
	}

	masked := make([]string, len(values))

	for i := range values {
		masked[i] = headerMask
	}

	return masked
}
