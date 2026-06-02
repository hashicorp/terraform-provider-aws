// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package kinesis_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccKinesis_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"AccountSettings": {
			acctest.CtBasic: testAccAccountSettings_basic,
			"enabled":       testAccAccountSettings_enabled,
			"Identity":      testAccKinesisAccountSettings_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
