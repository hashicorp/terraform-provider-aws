package route53recoverycontrolconfig_test

import (
	"testing"
)

func TestAccAWSRoute53RecoveryControlConfig_serial(t *testing.T) {
	// These tests are only non-parallel because of low quota limits.
	// ServiceQuotaExceededException: AwsAccountId(X) has 2 Meridian clusters. Limit 2
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			"basic":      testAccCluster_basic,
			"disappears": testAccCluster_disappears,
		},
		"ControlPanel": {
			"basic":      testAccControlPanel_basic,
			"disappears": testAccControlPanel_disappears,
		},
		"RoutingControl": {
			"basic":                 testAccRoutingControl_basic,
			"disappears":            testAccRoutingControl_disappears,
			"nonDefaultControlPane": testAccRoutingControl_nonDefaultControlPanel,
		},
		"SafetyRule": {
			"assertionRule": testAccSafetyRule_assertionRule,
			"gatingRule":    testAccSafetyRule_gatingRule,
			"disappears":    testAccSafetyRule_disappears,
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
