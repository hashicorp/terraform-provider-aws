// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package chime_test

import (
	"testing"

	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccChime_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]map[string]func(t *testing.T){
		"VoiceConnector": {
			acctest.CtBasic: testAccVoiceConnector_basic,
			"disappears":    testAccVoiceConnector_disappears,
			"update":        testAccVoiceConnector_update,
			"tags":          testAccVoiceConnector_tags,
		},
		"VoiceConnectorGroup": {
			acctest.CtBasic: testAccVoiceConnectorGroup_basic,
			"disappears":    testAccVoiceConnectorGroup_disappears,
			"update":        testAccVoiceConnectorGroup_update,
		},
		"VoiceConnectorLogging": {
			acctest.CtBasic: testAccVoiceConnectorLogging_basic,
			"disappears":    testAccVoiceConnectorLogging_disappears,
			"update":        testAccVoiceConnectorLogging_update,
		},
		"VoiceConnectorOrigination": {
			acctest.CtBasic: testAccVoiceConnectorOrigination_basic,
			"disappears":    testAccVoiceConnectorOrigination_disappears,
			"update":        testAccVoiceConnectorOrigination_update,
		},
		"VoiceConnectorStreaming": {
			acctest.CtBasic: testAccVoiceConnectorStreaming_basic,
			"disappears":    testAccVoiceConnectorStreaming_disappears,
			"update":        testAccVoiceConnectorStreaming_update,
		},
		"VoiceConnectorTermination": {
			acctest.CtBasic: testAccVoiceConnectorTermination_basic,
			"disappears":    testAccVoiceConnectorTermination_disappears,
			"update":        testAccVoiceConnectorTermination_update,
		},
		"VoiceConnectorTerminationCredentials": {
			acctest.CtBasic: testAccVoiceConnectorTerminationCredentials_basic,
			"disappears":    testAccVoiceConnectorTerminationCredentials_disappears,
			"update":        testAccVoiceConnectorTerminationCredentials_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
