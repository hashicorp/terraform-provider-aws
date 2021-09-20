package aws

import (
	"testing"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/provider"
)

func TestAccAWSRoute53RecoveryControlConfig_serial(t *testing.T) {
	// These tests are only non-parallel because of low quota limits.
	// ServiceQuotaExceededException: AwsAccountId(X) has 2 Meridian clusters. Limit 2
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			"basic":      testAccAWSRoute53RecoveryControlConfigCluster_basic,
			"disappears": testAccAWSRoute53RecoveryControlConfigCluster_disappears,
		},
		"ControlPanel": {
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
