// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package types

import (
	"github.com/YakDriver/regexache"
)

// servicePrincipalRegexp matches AWS service principal names.
// Service principals follow the pattern: service-id.amazonaws.com, service-id.amazon.com, or service-id.amazonaws.com.cn.
// Examples: ec2.amazonaws.com, s3.amazonaws.com, elasticmapreduce.amazonaws.com, s3.amazonaws.com.cn.
var servicePrincipalRegexp = regexache.MustCompile(`^([0-9a-z-]+\.){1,4}(amazonaws|amazon)\.(com|com\.cn)$`)

// IsServicePrincipal returns whether or not the specified string is a valid AWS service principal.
func IsServicePrincipal(s string) bool {
	return servicePrincipalRegexp.MatchString(s)
}
