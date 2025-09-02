// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fwdiag

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/YakDriver/regexache"
)

func ExpectAttributeRequiredWhenError(neededPath, otherPath, value string) *regexp.Regexp {
	return regexache.MustCompile(strings.ReplaceAll(fmt.Sprintf(`Attribute "%s" must be specified when "%s" is "%s".`,
		regexp.QuoteMeta(neededPath),
		regexp.QuoteMeta(otherPath),
		value,
	), " ", `(\n?\s+)`))
}

func ExpectAttributeConflictsWhenError(path, otherPath, value string) *regexp.Regexp {
	return regexache.MustCompile(strings.ReplaceAll(fmt.Sprintf(`Attribute "%s" cannot be specified when "%s" is "%s".`,
		regexp.QuoteMeta(path),
		regexp.QuoteMeta(otherPath),
		value,
	), " ", `(\n?\s+)`))
}
