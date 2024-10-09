// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.EC2ServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"VolumeTypeNotAvailableInRegion",
		"Invalid value specified for Phase",
		"You have reached the maximum allowed number of license configurations created in one day",
		"specified zone does not support multi-attach-enabled volumes",
		"Unsupported volume type",
		"HostLimitExceeded",
		"ReservationCapacityExceeded",
		"InsufficientInstanceCapacity",
		"There is no Spot capacity available that matches your request",
	)
}
