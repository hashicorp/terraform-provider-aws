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

	"github.com/aws/aws-sdk-go-v2/aws"
	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

func cleanDelegationSetID(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}

// normalizeDomainNameToAPI converts an user input into the Route 53 API's domain name representation.
// See https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones.
// See https://datatracker.ietf.org/doc/html/rfc4343#section-2.1.
func normalizeDomainNameToAPI(input string) string {
	var output strings.Builder
	br := bufio.NewReader(strings.NewReader(input))

	for {
		ch, _, err := br.ReadRune()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return ""
		}

		switch {
		case ch >= 'A' && ch <= 'Z':
			output.WriteRune(unicode.ToLower(ch))
		case ch >= '0' && ch <= '9' || ch >= 'a' && ch <= 'z' || ch == '-' || ch == '.' || ch == '_':
			output.WriteRune(ch)
		case ch == '\\':
			const (
				lenOctalCode = 3
			)
			if bytes, err := br.Peek(lenOctalCode); err == nil && tfslices.All(bytes, func(b byte) bool {
				return b >= '0' && b <= '7' // Octal.
			}) {
				output.WriteRune(ch)
				output.WriteString(string(bytes))
				_, _ = br.Discard(lenOctalCode)
				continue
			}
			fallthrough
		default:
			// Three-digit octal code.
			output.WriteString(fmt.Sprintf("\\%03o", ch))
		}
	}

	return output.String()
}

// cleanZoneID is used to remove the leading "/hostedzone/" from a hosted zone ID.
func cleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}

func normalizeAliasDomainName(v interface{}) string {
	var s string
	switch v := v.(type) {
	case *string:
		s = aws.ToString(v)
	case string:
		s = v
	default:
		return ""
	}

	return normalizeDomainNameToAPI(strings.TrimSuffix(s, "."))
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

	return normalizeDomainNameToAPI(strings.TrimSuffix(s, "."))
}

// fqdn appends a single dot (".") to the input string if necessary.
func fqdn(name string) string {
	if n := len(name); n == 0 || name[n-1] == '.' {
		return name
	} else {
		return name + "."
	}
}
