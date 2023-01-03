package sqs_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	tfsqs "github.com/hashicorp/terraform-provider-aws/internal/service/sqs"
)

func TestQueuePolicyMigrateState(t *testing.T) {
	cases := map[string]struct {
		StateVersion int
		ID           string
		Attributes   map[string]string
		Expected     string
		Meta         interface{}
	}{
		"v0_1": {
			StateVersion: 0,
			ID:           "sqs-policy-https://queue.amazonaws.com/0123456789012/myqueue",
			Attributes: map[string]string{
				"policy":    "{}",
				"queue_url": "https://queue.amazonaws.com/0123456789012/myqueue",
			},
			Expected: "https://queue.amazonaws.com/0123456789012/myqueue",
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         tc.ID,
			Attributes: tc.Attributes,
		}
		is, err := tfsqs.QueuePolicyMigrateState(
			tc.StateVersion, is, tc.Meta)

		if err != nil {
			t.Fatalf("bad: %s, err: %#v", tn, err)
		}

		if is.ID != tc.Expected {
			t.Fatalf("bad sqs queue policy id: %s\n\n expected: %s", is.ID, tc.Expected)
		}
	}
}
