// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"fmt"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
)

// randomFrameworkName returns a resource name that matches the pattern for Framework names
func randomFrameworkName() string {
	return fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
}

// randomReportPlanName returns a resource name that matches the pattern for Report Plan names
func randomReportPlanName() string {
	return fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
}
