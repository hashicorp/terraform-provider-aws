// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// CanonicalAccountIDPatternNoAnchors is the canonical regex pattern for validating AWS account IDs,
// without anchors for use within larger regex patterns.
// See https://docs.aws.amazon.com/accounts/latest/reference/manage-acct-identifiers.html.
// Uses non-capturing groups to avoid interfering with capture groups in composed patterns.
const CanonicalAccountIDPatternNoAnchors = `\d{12}`

// CanonicalAccountIDPattern is the anchored version of CanonicalAccountIDPatternNoAnchors.
const CanonicalAccountIDPattern = `^` + CanonicalAccountIDPatternNoAnchors + `$`

// IsAWSAccountID returns whether or not the specified string is a valid AWS account ID.
func IsAWSAccountID(s string) bool { // nosemgrep:ci.aws-in-func-name
	return regexache.MustCompile(CanonicalAccountIDPattern).MatchString(s)
}
