package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAwsCloudFrontDistributionMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			Attributes: map[string]string{
				"wait_for_deployment": "",
			},
			Expected: map[string]string{
				"wait_for_deployment": "true",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		tfResource := resourceAwsCloudFrontDistribution()

		if tfResource.MigrateState == nil {
			t.Fatalf("bad: %s, err: missing MigrateState function in resource", tn)
		}

		is, err := tfResource.MigrateState(tc.StateVersion, is, tc.Meta)
		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		for k, v := range tc.Expected {
			if is.Attributes[k] != v {
				t.Fatalf(
					"bad: %s\n\n expected: %#v -> %#v\n got: %#v -> %#v\n in: %#v",
					tn, k, v, k, is.Attributes[k], is.Attributes)
			}
		}
	}
}
