// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"bufio"
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

// normalizeNameIntoRoute53APIRepresentation converts an user input into Route53's record name representation.
// See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones
// See https://datatracker.ietf.org/doc/html/rfc4343#section-2.1
func normalizeNameIntoRoute53APIRepresentation(input string) string {
	var ret string

	br := bufio.NewReader(strings.NewReader(input))

	for {
		ch, _, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return ""
		}

		ch = unicode.ToLower(ch)

		// when backslack is found, check if a beginning of escape code
		switch {
		case ch == '\\':
			esc, err := br.Peek(3)
			if err == nil {
				if isAllNumeric(string(esc)) {
					ret += "\\" + string(esc)

					// advanced 3 bytes
					_, _ = br.Discard(3)
					continue
				}
			}
			// treat it as "\"  and carry on
			ret += "\\"
		case needRoute53EscapeCode(ch):
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

// needRoute53EscapeCode returns true if a given rune needs an escape code for Route53 representation.
// https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones
// > If the domain name includes any characters other than a to z, 0 to 9, - (hyphen), or _ (underscore), Route 53 API actions return the characters as escape codes
func needRoute53EscapeCode(r rune) bool {
	return !regexache.MustCompile(`[0-9A-Za-z_\-.]`).MatchString(string(r))
}

// CleanZoneID is used to remove the leading /hostedzone/
func cleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}

func normalizeAliasName(alias interface{}) string {
	output := strings.ToLower(alias.(string))
	return normalizeNameIntoRoute53APIRepresentation(strings.TrimSuffix(output, "."))
}

// normalizeZoneName is used to remove the trailing period
// and apply consistent casing to "name" or "domain name"
// attributes returned from the Route53 API or provided as
// user input.
//
// The single dot (".") domain name is returned as-is.
// Uppercase letters are converted to lowercase.
func normalizeZoneName(v interface{}) string {
	var str string
	switch value := v.(type) {
	case *string:
		str = aws.ToString(value)
	case string:
		str = value
	default:
		return ""
	}

	if str == "." {
		return str
	}

	return normalizeNameIntoRoute53APIRepresentation(
		strings.ToLower(strings.TrimSuffix(str, ".")),
	)
}

func fqdn(name string) string {
	n := len(name)
	if n == 0 || name[n-1] == '.' {
		return name
	} else {
		return name + "."
	}
}
