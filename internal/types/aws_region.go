// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// CanonicalRegionPatternNoAnchors is the canonical regex pattern for validating AWS region names,
// without anchors for use within larger regex patterns.
// It supports standard regions (us-east-1), government regions (us-gov-west-1),
// and sovereign cloud regions (eusc-de-east-1).
// Uses non-capturing groups to avoid interfering with capture groups in composed patterns.
const CanonicalRegionPatternNoAnchors = `[a-z]{2,4}(?:-[a-z]+)+-\d{1,2}`

// CanonicalRegionPattern is the anchored version of CanonicalRegionPatternNoAnchors.
const CanonicalRegionPattern = `^` + CanonicalRegionPatternNoAnchors + `$`

// IsAWSRegion returns whether or not the specified string is a valid AWS Region.
func IsAWSRegion(s string) bool { // nosemgrep:ci.aws-in-func-name
	return regexache.MustCompile(CanonicalRegionPattern).MatchString(s)
}
