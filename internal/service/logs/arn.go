// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"strings"
)

const (
	logGroupARNWildcardSuffix = ":*"
)

// trimLogGroupARNWildcardSuffix trims any wilcard suffix from a Log Group ARN.
func trimLogGroupARNWildcardSuffix(arn string) string {
	return strings.TrimSuffix(arn, logGroupARNWildcardSuffix)
}
