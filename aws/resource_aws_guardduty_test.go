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
		"IPSet": {
			"basic":  testAccAwsGuardDutyIpset_basic,
			"import": testAccAwsGuardDutyIpset_import,
		},
		"ThreatIntelSet": {
			"basic":  testAccAwsGuardDutyThreatintelset_basic,
			"import": testAccAwsGuardDutyThreatintelset_import,
		},
		"Member": {
			"basic":  testAccAwsGuardDutyMember_basic,
			"invite": testAccAwsGuardDutyMember_invite,
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
