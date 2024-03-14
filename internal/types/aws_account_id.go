// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// IsAWSAccountID returns whether or not the specified string is a valid AWS account ID.
func IsAWSAccountID(s string) bool {
	return regexache.MustCompile(`^\d{12}$`).MatchString(s)
}
