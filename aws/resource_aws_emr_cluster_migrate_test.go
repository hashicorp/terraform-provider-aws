package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAwsEMRClusterMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"change `ebs_config` from set to list": {
			StateVersion: 0,
			Attributes: map[string]string{
				"instance_group.#":                                     "1",
				"instance_group.1693978638.ebs_config.#":               "2",
				"instance_group.1693978638.ebs_config.1693978638.name": "test1",
				"instance_group.1693978638.ebs_config.1693978638.size": "10",
				"instance_group.1693978638.ebs_config.112349887.name":  "test2",
				"instance_group.1693978638.ebs_config.112349887.size":  "20",
			},
			Expected: map[string]string{
				"instance_group.#":                            "1",
				"instance_group.1693978638.ebs_config.#":      "2",
				"instance_group.1693978638.ebs_config.0.name": "test1",
				"instance_group.1693978638.ebs_config.0.size": "10",
				"instance_group.1693978638.ebs_config.1.name": "test2",
				"instance_group.1693978638.ebs_config.1.size": "20",
			},
		},
	}
	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsEMRClusterMigrateState(
			tc.StateVersion, is, tc.Meta)

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

func TestAwsEMRClusterMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsEMRClusterMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = resourceAwsEMRClusterMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
