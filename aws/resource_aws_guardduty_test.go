package aws

import (
	"testing"
)

func TestAccAWSGuardDuty(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Detector": {
			"basic":  testAccAwsGuardDutyDetector_basic,
			"import": testAccAwsGuardDutyDetector_import,
		},
		"Member": {
			"basic":  testAccAwsGuardDutyMember_basic,
			"import": testAccAwsGuardDutyMember_import,
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
