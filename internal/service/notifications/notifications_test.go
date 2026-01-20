// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package notifications_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccNotifications_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"OrganizationsAccess": {
			acctest.CtBasic: testAccOrganizationsAccess_basic,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
