// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"
	"regexp"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validCustomAMIID(v interface{}, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 256 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 256 characters", k))
	}

	if !regexp.MustCompile(`^ami\-[a-z0-9]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must begin with 'ami-' and be comprised of only [a-z0-9]: %v", k, value))
	}

	return
}

func validEBSVolumeType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"gp3",
		"gp2",
		"io1",
		"standard",
		"st1",
		"sc1",
	}, false)
}
