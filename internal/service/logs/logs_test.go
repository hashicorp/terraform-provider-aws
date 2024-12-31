// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package logs_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccLogs_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DeliverySource": {
			acctest.CtBasic:      testAccDeliverySource_basic,
			acctest.CtDisappears: testAccDeliverySource_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
