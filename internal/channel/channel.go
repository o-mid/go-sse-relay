// Package channel validates stream and hook channel names.
package channel

import "unicode/utf8"

const maxLen = 128

// Valid reports whether name is safe for URL path and query use.
// Allowed: letters, digits, '_', '-', '.', ':'.
func Valid(name string) bool {
	if name == "" || utf8.RuneCountInString(name) > maxLen {
		return false
	}
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.' || r == ':':
		default:
			return false
		}
	}
	return true
}
