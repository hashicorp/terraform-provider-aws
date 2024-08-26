// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package quicksight_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccQuickSight_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AccountSubscription": {
			acctest.CtBasic:      testAccAccountSubscription_basic,
			acctest.CtDisappears: testAccAccountSubscription_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
