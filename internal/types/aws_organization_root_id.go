// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import "github.com/YakDriver/regexache"

// IsAWSOrganizationRootID returns whether or not the specified string is a valid AWS Organization Root ID.
func IsAWSOrganizationRootID(s string) bool { // nosemgrep:ci.aws-in-func-name
	return regexache.MustCompile(`^r-[0-9a-z]{4,32}$`).MatchString(s)
}
