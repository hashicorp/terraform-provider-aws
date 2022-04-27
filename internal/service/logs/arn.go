package logs

import (
	"strings"
)

const (
	logGroupARNWildcardSuffix = ":*"
)

// TrimLogGroupARNWildcardSuffix trims any wilcard suffix from a Log Group ARN.
func TrimLogGroupARNWildcardSuffix(arn string) string {
	return strings.TrimSuffix(arn, logGroupARNWildcardSuffix)
}
