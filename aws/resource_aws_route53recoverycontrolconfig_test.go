package aws

import (
	"testing"
)

func TestAccAWSRoute53RecoveryControlConfig_serial(t *testing.T) {
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			// if more than 2, increase "Meridian clusters" quota or switch these to non-parallel tests
			"basic":      testAccAWSRoute53RecoveryControlConfigCluster_basic,
			"disappears": testAccAWSRoute53RecoveryControlConfigCluster_disappears,
		},
		"ControlPanel": {
			// if more than 2, increase "Meridian clusters" quota or switch these to non-parallel tests
			"basic":      testAccAWSRoute53RecoveryControlConfigControlPanel_basic,
			"disappears": testAccAWSRoute53RecoveryControlConfigControlPanel_disappears,
		},
		"RoutingControl": {
			"basic":                 testAccAWSRoute53RecoveryControlConfigRoutingControl_basic,
			"disappears":            testAccAWSRoute53RecoveryControlConfigRoutingControl_disappears,
			"nonDefaultControlPane": testAccAWSRoute53RecoveryControlConfigRoutingControl_nonDefaultControlPanel,
		},
		"SafetyRule": {
			"assertionRule": testAccAWSRoute53RecoveryControlConfigSafetyRule_assertionRule,
			"gatingRule":    testAccAWSRoute53RecoveryControlConfigSafetyRule_gatingRule,
			"disappears":    testAccAWSRoute53RecoveryControlConfigSafetyRule_disappears,
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
