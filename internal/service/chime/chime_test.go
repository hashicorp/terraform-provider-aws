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
			"basic":      testAccVoiceConnector_basic,
			"disappears": testAccVoiceConnector_disappears,
			"update":     testAccVoiceConnector_update,
			"tags":       testAccVoiceConnector_tags,
		},
		"VoiceConnectorGroup": {
			"basic":      testAccVoiceConnectorGroup_basic,
			"disappears": testAccVoiceConnectorGroup_disappears,
			"update":     testAccVoiceConnectorGroup_update,
		},
		"VoiceConnectorLogging": {
			"basic":      testAccVoiceConnectorLogging_basic,
			"disappears": testAccVoiceConnectorLogging_disappears,
			"update":     testAccVoiceConnectorLogging_update,
		},
		"VoiceConnectorOrigination": {
			"basic":      testAccVoiceConnectorOrigination_basic,
			"disappears": testAccVoiceConnectorOrigination_disappears,
			"update":     testAccVoiceConnectorOrigination_update,
		},
		"VoiceConnectorStreaming": {
			"basic":      testAccVoiceConnectorStreaming_basic,
			"disappears": testAccVoiceConnectorStreaming_disappears,
			"update":     testAccVoiceConnectorStreaming_update,
		},
		"VoiceConnectorTermination": {
			"basic":      testAccVoiceConnectorTermination_basic,
			"disappears": testAccVoiceConnectorTermination_disappears,
			"update":     testAccVoiceConnectorTermination_update,
		},
		"VoiceConnectorTerminationCredentials": {
			"basic":      testAccVoiceConnectorTerminationCredentials_basic,
			"disappears": testAccVoiceConnectorTerminationCredentials_disappears,
			"update":     testAccVoiceConnectorTerminationCredentials_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
