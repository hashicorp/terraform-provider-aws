// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ssm_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

// These tests affect regional defaults, so they needs to be serialized
func TestAccSSM_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"DefaultPatchBaseline": {
			acctest.CtBasic:        testAccSSMDefaultPatchBaseline_basic,
			acctest.CtDisappears:   testAccSSMDefaultPatchBaseline_disappears,
			"otherOperatingSystem": testAccSSMDefaultPatchBaseline_otherOperatingSystem,
			"patchBaselineARN":     testAccSSMDefaultPatchBaseline_patchBaselineARN,
			"systemDefault":        testAccSSMDefaultPatchBaseline_systemDefault,
			"update":               testAccSSMDefaultPatchBaseline_update,
			"deleteDefault":        testAccSSMPatchBaseline_deleteDefault,
			"multiRegion":          testAccSSMDefaultPatchBaseline_multiRegion,
			"wrongOperatingSystem": testAccSSMDefaultPatchBaseline_wrongOperatingSystem,
		},
		"PatchBaseline": {
			"deleteDefault": testAccSSMPatchBaseline_deleteDefault,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
