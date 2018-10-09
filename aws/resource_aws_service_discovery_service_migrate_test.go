package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAwsServiceDiscoveryServiceMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			ID:           "tf-testing-file",
			Attributes: map[string]string{
				"name":                            "test-name",
				"dns_config.#":                    "1",
				"dns_config.0.namespace_id":       "test-namespace",
				"dns_config.0.dns_records.#":      "1",
				"dns_config.0.dns_records.0.ttl":  "10",
				"dns_config.0.dns_records.0.type": "A",
				"arn":                             "arn",
			},
			Expected: map[string]string{
				"name":                            "test-name",
				"dns_config.#":                    "1",
				"dns_config.0.namespace_id":       "test-namespace",
				"dns_config.0.routing_policy":     "MULTIVALUE",
				"dns_config.0.dns_records.#":      "1",
				"dns_config.0.dns_records.0.ttl":  "10",
				"dns_config.0.dns_records.0.type": "A",
				"arn":                             "arn",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsServiceDiscoveryServiceMigrateState(tc.StateVersion, is, tc.Meta)

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
