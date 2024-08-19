// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package cloudfront_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfcloudfront "github.com/hashicorp/terraform-provider-aws/internal/service/cloudfront"
)

func TestDistributionMigrateState(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_to_v1": {
			StateVersion: 0,
			Attributes: map[string]string{
				"wait_for_deployment": "",
			},
			Expected: map[string]string{
				"wait_for_deployment": acctest.CtTrue,
			},
		},
	}

	for testName, testCase := range testCases {
		instanceState := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: testCase.Attributes,
		}

		tfResource := tfcloudfront.ResourceDistribution()

		if tfResource.MigrateState == nil {
			t.Fatalf("bad: %s, err: missing MigrateState function in resource", testName)
		}

		instanceState, err := tfResource.MigrateState(testCase.StateVersion, instanceState, testCase.Meta)
		if err != nil {
			t.Fatalf("bad: %s, err: %#v", testName, err)
		}

		for key, expectedValue := range testCase.Expected {
			if instanceState.Attributes[key] != expectedValue {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					testName, key, expectedValue, key, instanceState.Attributes[key], instanceState.Attributes)
			}
		}
	}
}
