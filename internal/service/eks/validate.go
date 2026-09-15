// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package eks

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validClusterName(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) < 1 || len(value) > 100 {
		errors = append(errors, fmt.Errorf(
			"%q length must be between 1-100 characters: %q", k, value))
	}

	// https://docs.aws.amazon.com/eks/latest/APIReference/API_CreateCluster.html#API_CreateCluster_RequestSyntax
	pattern := `^[0-9A-Za-z][0-9A-Za-z_-]*$`
	if !regexache.MustCompile(pattern).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q doesn't comply with restrictions (%q): %q",
			k, pattern, value))
	}

	return
}

var ipV4CIDRString = `10|172\.(1[6-9]|2[0-9]|3[0-1])|192\.168`
var cgNATString = `100\.(6[4-9]|7[0-9]|8[0-9]|9[0-9]|10[0-9]|11[0-9]|12[0-7])`
var regexString = fmt.Sprintf(`^(%s|%s)\..*`, ipV4CIDRString, cgNATString)
var validateIPv4CIDRPrivateRange = validation.StringMatch(regexache.MustCompile(regexString), "must be within 10.0.0.0/8, 172.16.0.0/12, 192.168.0.0/16 or 100.64.0.0/10")
