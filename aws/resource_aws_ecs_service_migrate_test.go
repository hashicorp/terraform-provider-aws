package aws

import (
	"testing"

	"github.com/hashicorp/terraform/terraform"
)

func TestAwsEcsServiceMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         interface{}
	}{
		"hoge": {
			StateVersion: 0,
			Attributes: map[string]string{
				"placement_strategy.#":                "2",
				"placement_strategy.2750134989.field": "instanceId",
				"placement_strategy.2750134989.type":  "spread",
				"placement_strategy.3619322362.field": "attribute:ecs.availability-zone",
				"placement_strategy.3619322362.type":  "spread",
			},
			Expected: map[string]string{
				"placement_strategy.#":       "2",
				"placement_strategy.0.field": "instanceId",
				"placement_strategy.0.type":  "spread",
				"placement_strategy.1.field": "attribute:ecs.availability-zone",
				"placement_strategy.1.type":  "spread",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := resourceAwsEcsServiceMigrateState(tc.StateVersion, is, tc.Meta)

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

func TestAwsEcsServiceMigrateState_empty(t *testing.T) {
	var is *terraform.InstanceState
	var meta interface{}

	// should handle nil
	is, err := resourceAwsEcsServiceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	is, err = resourceAwsEcsServiceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}
