// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs

import (
	"errors"
	"strings"
)

const (
	logGroupARNWildcardSuffix = ":*"
)

// TrimLogGroupARNWildcardSuffix trims any wilcard suffix from a Log Group ARN.
func TrimLogGroupARNWildcardSuffix(arn string) string {
	return strings.TrimSuffix(arn, logGroupARNWildcardSuffix)
}

func logGroupArnToName(arn string) (string, error) {
	parts := strings.SplitN(arn, "log-group:", 2)
	if len(parts) > 1 {
		return parts[1], nil
	}
	return "", errors.New("invalid log group ARN")
}
