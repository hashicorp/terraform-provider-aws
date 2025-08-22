// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package route53recoverycontrolconfig_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRoute53RecoveryControlConfig_serial(t *testing.T) {
	t.Parallel()

	// These tests are only non-parallel because of low quota limits.
	// ServiceQuotaExceededException: AwsAccountId(X) has 2 Meridian clusters. Limit 2
	testCases := map[string]map[string]func(t *testing.T){
		"Cluster": {
			acctest.CtBasic:      testAccCluster_basic,
			acctest.CtDisappears: testAccCluster_disappears,
		},
		"ControlPanel": {
			acctest.CtBasic:      testAccControlPanel_basic,
			acctest.CtDisappears: testAccControlPanel_disappears,
		},
		"RoutingControl": {
			acctest.CtBasic:         testAccRoutingControl_basic,
			acctest.CtDisappears:    testAccRoutingControl_disappears,
			"nonDefaultControlPane": testAccRoutingControl_nonDefaultControlPanel,
		},
		"SafetyRule": {
			"assertionRule":      testAccSafetyRule_assertionRule,
			"gatingRule":         testAccSafetyRule_gatingRule,
			acctest.CtDisappears: testAccSafetyRule_disappears,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
