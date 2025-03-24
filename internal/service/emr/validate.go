// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package emr

import (
	"fmt"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func validCustomAMIID(v any, k string) (ws []string, errors []error) {
	value := v.(string)
	if len(value) > 256 {
		errors = append(errors, fmt.Errorf("%q cannot be longer than 256 characters", k))
	}

	if !regexache.MustCompile(`^ami\-[0-9a-z]+$`).MatchString(value) {
		errors = append(errors, fmt.Errorf(
			"%q must begin with 'ami-' and be comprised of only [0-9a-z]: %v", k, value))
	}

	return
}

func validEBSVolumeType() schema.SchemaValidateFunc {
	return validation.StringInSlice([]string{
		"gp3",
		"gp2",
		"io1",
		"io2",
		"standard",
		"st1",
		"sc1",
	}, false)
}
