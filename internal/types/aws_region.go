// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// IsAWSRegion returns whether or not the specified string is a valid AWS Region.
func IsAWSRegion(s string) bool { // nosemgrep:ci.aws-in-func-name
	return regexache.MustCompile(`^[a-z]{2}(-[a-z]+)+-\d{1,2}$`).MatchString(s)
}
