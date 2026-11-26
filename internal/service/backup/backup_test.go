// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package backup_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// randomFrameworkName returns a resource name that matches the pattern for Framework names
func randomFrameworkName(t *testing.T) string {
	return fmt.Sprintf("tf_acc_test_%s", acctest.RandString(t, 7))
}

// randomReportPlanName returns a resource name that matches the pattern for Report Plan names
func randomReportPlanName(t *testing.T) string {
	return fmt.Sprintf("tf_acc_test_%s", acctest.RandString(t, 7))
}

// randomTieringConfigurationName returns a resource name that matches the pattern for Tiering Configuration names
func randomTieringConfigurationName() string {
	return fmt.Sprintf("tf_acc_test_%s", sdkacctest.RandString(7))
}
