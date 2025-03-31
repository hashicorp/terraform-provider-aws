// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package dns

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"
	"unicode"

	tfslices "github.com/hashicorp/terraform-provider-aws/internal/slices"
)

// normalizeCasingAndEscapeCodes handles matching the casing and escape code representations
// used by the Route53 API
//
// Ref:
// - https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones.
// - https://datatracker.ietf.org/doc/html/rfc4343#section-2.1.
func normalizeCasingAndEscapeCodes(input string) string {
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

// Normalize is used to remove the trailing period and apply consistent
// casing to Route53 domain names
//
// This can be applied to hosted zone names and record names returned from the Route53
// API, or provided via a Terraform configuration. The single dot (".") domain name is
// returned as-is.
//
// Ref:
// - https://docs.aws.amazon.com/Route53/latest/DeveloperGuide/DomainNameFormat.html#domain-name-format-hosted-zones.
func Normalize(s string) string {
	if s == "." {
		return s
	}
	return normalizeCasingAndEscapeCodes(strings.TrimSuffix(s, "."))
}
