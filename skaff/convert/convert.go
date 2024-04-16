// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package convert

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/YakDriver/regexache"
)

// ToSnakeCase converts a camel cased string to snake case
//
// If the override argument is a non-empty string, its value is returned
// unchanged.
func ToSnakeCase(upper string, override string) string {
	if override != "" {
		return override
	}

	re := regexache.MustCompile(`([a-z])([A-Z]{2,})`)
	upper = re.ReplaceAllString(upper, `${1}_${2}`)

	re2 := regexache.MustCompile(`([A-Z][a-z])`)
	return strings.TrimPrefix(strings.ToLower(re2.ReplaceAllString(upper, `_$1`)), "_")
}

// ToHumanResName converts a camel cased string to a human readable name
func ToHumanResName(upper string) string {
	re := regexache.MustCompile(`([a-z])([A-Z]{2,})`)
	upper = re.ReplaceAllString(upper, `${1} ${2}`)

	re2 := regexache.MustCompile(`([A-Z][a-z])`)
	return strings.TrimPrefix(re2.ReplaceAllString(upper, ` $1`), " ")
}

// ToProviderResourceName adds the appropriate prefix to a snake cased name
// of a resource or data source
func ToProviderResourceName(servicePackage, snakeName string) string {
	return fmt.Sprintf("aws_%s_%s", servicePackage, snakeName)
}

// ToLowercasePrefix converts a string beginning with uppercase letters
// to begin lowercased
//
// Specifically, this is used to take user-provided input beginning with uppercase
// letters, and transform it so that it can be used to name a private struct. This
// function assumes all characters are unicode.
func ToLowercasePrefix(s string) string {
	var hasLower bool
	var splitIdx int
	for i, char := range s {
		if unicode.IsLower(char) {
			hasLower = true
			break
		}
		splitIdx = i
	}

	if !hasLower {
		return strings.ToLower(s)
	}
	if splitIdx == 0 && len(s) > 0 {
		splitIdx++
	}
	return strings.ToLower(s[:splitIdx]) + s[splitIdx:]
}
