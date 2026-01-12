// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// CanonicalRegionPattern is the canonical regex pattern for validating AWS region names.
// It supports standard regions (us-east-1), government regions (us-gov-west-1),
// and sovereign cloud regions (eusc-de-east-1).
const CanonicalRegionPattern = `^[a-z]{2,4}(-[a-z]+)+-\d{1,2}$`

// CanonicalRegionPatternNoAnchors is the same pattern without ^ and $ anchors,
// for use within larger regex patterns.
const CanonicalRegionPatternNoAnchors = `[a-z]{2,4}(-[a-z]+)+-\d{1,2}`

// IsAWSRegion returns whether or not the specified string is a valid AWS Region.
func IsAWSRegion(s string) bool { // nosemgrep:ci.aws-in-func-name
	return regexache.MustCompile(CanonicalRegionPattern).MatchString(s)
}
