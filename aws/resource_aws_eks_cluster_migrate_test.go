package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAWSEksClusterMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			Attributes: map[string]string{
				"enabled_cluster_log_types.#": "5",
				"enabled_cluster_log_types.0": "api",
				"enabled_cluster_log_types.1": "audit",
				"enabled_cluster_log_types.2": "authenticator",
				"enabled_cluster_log_types.3": "controllerManager",
				"enabled_cluster_log_types.4": "scheduler",
				"version":                     "1.11",
			},
			Expected: map[string]string{
				"enabled_cluster_log_types.#":          "5",
				"enabled_cluster_log_types.2902841359": "api",
				"enabled_cluster_log_types.2451111801": "audit",
				"enabled_cluster_log_types.1524904079": "authenticator",
				"enabled_cluster_log_types.1136278952": "controllerManager",
				"enabled_cluster_log_types.1178397720": "scheduler",
				"version":                              "1.11",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "some_id",
			Attributes: tc.Attributes,
		}

		is, err := resourceAwsEksCluster().MigrateState(tc.StateVersion, is, tc.Meta)
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
