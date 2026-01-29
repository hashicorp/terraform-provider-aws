// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package mediapackagevod_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMediaPackageVOD_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"PackagingGroup": {
			acctest.CtBasic:      testAccMediaPackageVODPackagingGroup_basic,
			"logging":            testAccMediaPackageVODPackagingGroup_logging,
			"authorization":      testAccMediaPackageVODPackagingGroup_authorization,
			acctest.CtDisappears: testAccMediaPackageVODPackagingGroup_disappears,
			"tags":               testAccMediaPackageVODPackagingGroup_tagsSerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
