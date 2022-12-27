package ssm_test

import "testing"

// These tests affect regional defaults, so they needs to be serialized
func TestAccSSM_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"DefaultPatchBaseline": {
			"basic":                testAccSSMDefaultPatchBaseline_basic,
			"disappears":           testAccSSMDefaultPatchBaseline_disappears,
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

	for group, m := range testCases {
		m := m
		t.Run(group, func(t *testing.T) {
			for name, tc := range m {
				tc := tc
				t.Run(name, func(t *testing.T) {
					tc(t)
				})
			}
		})
	}
}
