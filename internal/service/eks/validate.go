// Copyright (c) HashiCorp, Inc.
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

var validateIPv4CIDRPrivateRange = validation.StringMatch(regexache.MustCompile(`^(10|172\.(1[6-9]|2[0-9]|3[0-1])|192\.168)\..*`), "must be within 10.0.0.0/8, 172.16.0.0/12, or 192.168.0.0/16")
