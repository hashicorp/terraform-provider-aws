// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53

import (
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
)

func cleanDelegationSetID(id string) string {
	return strings.TrimPrefix(id, "/delegationset/")
}

// CleanChangeID is used to remove the leading /change/
func CleanChangeID(ID string) string {
	return strings.TrimPrefix(ID, "/change/")
}

// Route 53 stores certain characters with the octal equivalent in ASCII format.
// This function converts all of these characters back into the original character.
// E.g. "*" is stored as "\\052" and "@" as "\\100"
func CleanRecordName(name string) string {
	str := name
	s, err := strconv.Unquote(`"` + str + `"`)
	if err != nil {
		return str
	}
	return s
}

// CleanZoneID is used to remove the leading /hostedzone/
func CleanZoneID(ID string) string {
	return strings.TrimPrefix(ID, "/hostedzone/")
}

func NormalizeAliasName(alias interface{}) string {
	output := strings.ToLower(alias.(string))
	return CleanRecordName(strings.TrimSuffix(output, "."))
}

// NormalizeZoneName is used to remove the trailing period
// and apply consistent casing to "name" or "domain name"
// attributes returned from the Route53 API or provided as
// user input.
//
// The single dot (".") domain name is returned as-is.
// Uppercase letters are converted to lowercase.
func NormalizeZoneName(v interface{}) string {
	var str string
	switch value := v.(type) {
	case *string:
		str = aws.ToString(value)
	case string:
		str = value
	default:
		return ""
	}

	if str == "." {
		return str
	}

	return strings.ToLower(strings.TrimSuffix(str, "."))
}
