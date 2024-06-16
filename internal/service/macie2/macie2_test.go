// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package macie2_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccMacie2_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"CustomDataIdentifier": {
			acctest.CtBasic:      testAccCustomDataIdentifier_basic,
			"name_generated":     testAccCustomDataIdentifier_Name_Generated,
			acctest.CtDisappears: testAccCustomDataIdentifier_disappears,
			"name_prefix":        testAccCustomDataIdentifier_NamePrefix,
			"classification_job": testAccCustomDataIdentifier_WithClassificationJob,
			"tags":               testAccCustomDataIdentifier_WithTags,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
