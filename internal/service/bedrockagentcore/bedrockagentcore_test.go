// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package bedrockagentcore_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccBedrockAgentCore_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"TokenVaultCMK": {
			acctest.CtBasic: testAccTokenVaultCMK_basic,
		},
		"PaymentManager": {
			acctest.CtBasic:       testAccPaymentManager_basic,
			acctest.CtDisappears:  testAccPaymentManager_disappears,
			names.AttrDescription: testAccPaymentManager_description,
			"tags":                testAccBedrockAgentCorePaymentManager_tagsSerial,
			"Identity":            testAccBedrockAgentCorePaymentManager_identitySerial,
		},
		"PaymentConnector": {
			acctest.CtBasic:       testAccPaymentConnector_basic,
			acctest.CtDisappears:  testAccPaymentConnector_disappears,
			names.AttrDescription: testAccPaymentConnector_description,
			"Identity":            testAccBedrockAgentCorePaymentConnector_identitySerial,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
