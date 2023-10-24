// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package s3control_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccS3ControlAccessGrants_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"Instance": {
			"basic":      testAccAccessGrantsInstance_basic,
			"disappears": testAccAccessGrantsInstance_disappears,
			"tags":       testAccAccessGrantsInstance_tags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
