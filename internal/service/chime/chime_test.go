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
			acctest.CtBasic:      testAccVoiceConnector_basic,
			acctest.CtDisappears: testAccVoiceConnector_disappears,
			"update":             testAccVoiceConnector_update,
			"tags":               testAccVoiceConnector_tags,
		},
		"VoiceConnectorGroup": {
			acctest.CtBasic:      testAccVoiceConnectorGroup_basic,
			acctest.CtDisappears: testAccVoiceConnectorGroup_disappears,
			"update":             testAccVoiceConnectorGroup_update,
		},
		"VoiceConnectorLogging": {
			acctest.CtBasic:      testAccVoiceConnectorLogging_basic,
			acctest.CtDisappears: testAccVoiceConnectorLogging_disappears,
			"update":             testAccVoiceConnectorLogging_update,
		},
		"VoiceConnectorOrigination": {
			acctest.CtBasic:      testAccVoiceConnectorOrigination_basic,
			acctest.CtDisappears: testAccVoiceConnectorOrigination_disappears,
			"update":             testAccVoiceConnectorOrigination_update,
		},
		"VoiceConnectorStreaming": {
			acctest.CtBasic:      testAccVoiceConnectorStreaming_basic,
			acctest.CtDisappears: testAccVoiceConnectorStreaming_disappears,
			"update":             testAccVoiceConnectorStreaming_update,
		},
		"VoiceConnectorTermination": {
			acctest.CtBasic:      testAccVoiceConnectorTermination_basic,
			acctest.CtDisappears: testAccVoiceConnectorTermination_disappears,
			"update":             testAccVoiceConnectorTermination_update,
		},
		"VoiceConnectorTerminationCredentials": {
			acctest.CtBasic:      testAccVoiceConnectorTerminationCredentials_basic,
			acctest.CtDisappears: testAccVoiceConnectorTerminationCredentials_disappears,
			"update":             testAccVoiceConnectorTerminationCredentials_update,
		},
	}

	acctest.RunSerialTests2Levels(t, testCases, 0)
}
