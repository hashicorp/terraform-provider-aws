// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// IsAWSAccountID returns whether or not the specified string is a valid AWS account ID.
func IsAWSAccountID(s string) bool { // nosemgrep:ci.aws-in-func-name
	// https://docs.aws.amazon.com/accounts/latest/reference/manage-acct-identifiers.html.
	return regexache.MustCompile(`^\d{12}$`).MatchString(s)
}
