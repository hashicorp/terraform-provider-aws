// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package elasticbeanstalk_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfelasticbeanstalk "github.com/hashicorp/terraform-provider-aws/internal/service/elasticbeanstalk"
)

func TestEnvironmentMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         any
	}{
		"v0_1_web": {
			StateVersion: 0,
			Attributes: map[string]string{
				"tier": "",
			},
			Expected: map[string]string{
				"tier": "WebServer",
			},
		},
		"v0_1_web_explicit": {
			StateVersion: 0,
			Attributes: map[string]string{
				"tier": "WebServer",
			},
			Expected: map[string]string{
				"tier": "WebServer",
			},
		},
		"v0_1_worker": {
			StateVersion: 0,
			Attributes: map[string]string{
				"tier": "Worker",
			},
			Expected: map[string]string{
				"tier": "Worker",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "e-abcde12345",
			Attributes: tc.Attributes,
		}
		_, err := tfelasticbeanstalk.EnvironmentMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}
	}
}
