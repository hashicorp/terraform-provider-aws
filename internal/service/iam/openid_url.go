package iam

import (
	"net/url"
	"strings"
)

// normalizeOpenIDURL normalizes OIDC issuer URLs.
// - trims spaces
// - if scheme-less, assumes https
func normalizeOpenIDURL(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return s
	}

	if !strings.Contains(s, "://") {
		s = "https://" + s
	}

	return s
}

// parseNormalizedOpenIDURL parses and returns the URL after normalization.
func parseNormalizedOpenIDURL(raw string) (*url.URL, error) {
	return url.Parse(normalizeOpenIDURL(raw))
}
