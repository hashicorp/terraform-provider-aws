// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/aws"
)

func cleanDelegationSetID(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}

// normalizeDomainNameToAPI converts an user input into the Route 53 API's domain name representation.
// See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones.
// See https://datatracker.ietf.org/doc/html/rfc4343#section-2.1.
func normalizeDomainNameToAPI(input string) string {
	var ret string

	br := bufio.NewReader(strings.NewReader(input))

	const lenEscapeCode = 3

	for {
		ch, _, err := br.ReadRune()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return ""
		}

		ch = unicode.ToLower(ch)

		// when backslack is found, check if a beginning of escape code
		switch {
		case ch == '\\':
			esc, err := br.Peek(lenEscapeCode)
			if err == nil {
				if isAllNumeric(string(esc)) {
					ret += "\\" + string(esc)

					// advanced 3 bytes
					_, _ = br.Discard(lenEscapeCode)
					continue
				}
			}
			// treat it as "\"  and carry on
			ret += "\\"
		case needEscapeCode(ch):
			// convert into escape code
			ret += fmt.Sprintf("\\%03o", ch)
		default:
			ret += string(ch)
		}
	}

	return ret
}

func isAllNumeric(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// needEscapeCode returns true if a given rune needs an escape code for Route53 representation.
// https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones
// > If the domain name includes any characters other than a to z, 0 to 9, - (hyphen), or _ (underscore), Route 53 API actions return the characters as escape codes
func needEscapeCode(r rune) bool {
	return !regexache.MustCompile(`[0-9A-Za-z_\-.]`).MatchString(string(r))
}

// cleanZoneID is used to remove the leading "/hostedzone/" from a hosted zone ID.
func cleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}

func normalizeAliasDomainName(alias interface{}) string {
	output := strings.ToLower(alias.(string))
	return normalizeDomainNameToAPI(strings.TrimSuffix(output, "."))
}

// normalizeDomainName is used to remove the trailing period and apply consistent casing to the domain names
// (including hosted zone names and record names) returned from the Route53 API or provided as configuration.
// The single dot (".") domain name is returned as-is.
// See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones.
func normalizeDomainName(v interface{}) string {
	var s string
	switch v := v.(type) {
	case *string:
		s = aws.ToString(v)
	case string:
		s = v
	default:
		return ""
	}

	if s == "." {
		return s
	}

	return normalizeDomainNameToAPI(
		strings.ToLower(strings.TrimSuffix(s, ".")),
	)
}

// fqdn appends a single dot (".") to the input string if necessary.
func fqdn(name string) string {
	if n := len(name); n == 0 || name[n-1] == '.' {
		return name
	} else {
		return name + "."
	}
}
