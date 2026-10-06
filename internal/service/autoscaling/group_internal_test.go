// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package autoscaling

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestExpandPutLifecycleHookInputOmitsUnsetOptionalFields(t *testing.T) {
	t.Parallel()

	// An optional argument the practitioner did not set arrives here as a zero
	// value. AWS validates these fields and has no reset value for them, so they
	// must be left out of the request entirely rather than sent as-is.
	input := expandPutLifecycleHookInput("test-asg", map[string]any{
		names.AttrName:            "test-hook",
		"lifecycle_transition":    "autoscaling:EC2_INSTANCE_LAUNCHING",
		"notification_metadata":   "",
		"notification_target_arn": "",
		names.AttrRoleARN:         "",
	})

	if input == nil {
		t.Fatal("expected lifecycle hook input")
	}

	for name, value := range map[string]*string{
		"NotificationMetadata":  input.NotificationMetadata,
		"NotificationTargetARN": input.NotificationTargetARN,
		"RoleARN":               input.RoleARN,
	} {
		if value != nil {
			t.Errorf("%s = %q, want omitted", name, aws.ToString(value))
		}
	}
}

func TestInitialLifecycleHookNames(t *testing.T) {
	t.Parallel()

	got := initialLifecycleHookNames([]any{
		map[string]any{names.AttrName: "second"},
		nil,
		map[string]any{names.AttrName: ""},
		map[string]any{names.AttrName: "first"},
	})
	want := []string{"first", "second"}

	if len(got) != len(want) {
		t.Fatalf("got %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("got[%d] = %q, want %q", i, got[i], want[i])
		}
	}
}

// TestInitialLifecycleHookIsNotForceNew locks in the core behavior of this
// resource: changing "initial_lifecycle_hook" must be reconciled in place
// rather than replacing the Auto Scaling Group.
func TestInitialLifecycleHookIsNotForceNew(t *testing.T) {
	t.Parallel()

	hook := resourceGroup().SchemaMap()["initial_lifecycle_hook"]
	if hook.ForceNew {
		t.Error("initial_lifecycle_hook must not be ForceNew")
	}

	for name, attr := range hook.Elem.(*schema.Resource).SchemaMap() {
		if attr.ForceNew {
			t.Errorf("initial_lifecycle_hook.%s must not be ForceNew", name)
		}
	}
}

// TestInitialLifecycleHookHasNoComputedAttributes guards against reintroducing
// Computed values into this set. Nested Optional attributes contribute to the
// set element hash, so populating AWS-side defaults for an attribute the
// practitioner omitted would change the element's identity and produce a
// permanent remove/add diff.
func TestInitialLifecycleHookHasNoComputedAttributes(t *testing.T) {
	t.Parallel()

	hook := resourceGroup().SchemaMap()["initial_lifecycle_hook"]

	for name, attr := range hook.Elem.(*schema.Resource).SchemaMap() {
		// "default_result" predates this behavior and is never written by Read.
		if name == "default_result" {
			continue
		}

		if attr.Computed {
			t.Errorf("initial_lifecycle_hook.%s must not be Computed", name)
		}
	}
}

// TestInitialLifecycleHookNeedsReplacement covers the one case PutLifecycleHook
// cannot express. Omitted arguments are left untouched on an existing hook and
// the three arguments below reject an empty string, so dropping one from the
// configuration is only possible by recreating the hook.
func TestInitialLifecycleHookNeedsReplacement(t *testing.T) {
	t.Parallel()

	const (
		metadata = "notification_metadata"
		target   = "notification_target_arn"
	)

	testCases := map[string]struct {
		oldHook, newHook map[string]any
		want             bool
	}{
		"notification_metadata removed": {
			oldHook: map[string]any{metadata: "payload"},
			newHook: map[string]any{metadata: ""},
			want:    true,
		},
		"notification_target_arn removed": {
			oldHook: map[string]any{target: "arn:aws:sns:us-west-2:123456789012:test"},
			newHook: map[string]any{target: ""},
			want:    true,
		},
		"role_arn removed": {
			oldHook: map[string]any{names.AttrRoleARN: "arn:aws:iam::123456789012:role/test"},
			newHook: map[string]any{names.AttrRoleARN: ""},
			want:    true,
		},
		"notification_metadata changed": {
			oldHook: map[string]any{metadata: "before"},
			newHook: map[string]any{metadata: "after"},
		},
		"notification_metadata added": {
			oldHook: map[string]any{metadata: ""},
			newHook: map[string]any{metadata: "payload"},
		},
		"unset argument stays unset": {
			oldHook: map[string]any{metadata: ""},
			newHook: map[string]any{metadata: ""},
		},
		// Both are always reported by AWS, defaulting to ABANDON and 3600, so a
		// removed argument cannot be told apart from an explicit one. Recreating
		// the hook would re-trigger on the AWS-assigned default forever.
		"heartbeat_timeout removed": {
			oldHook: map[string]any{"heartbeat_timeout": 300},
			newHook: map[string]any{"heartbeat_timeout": 0},
		},
		"default_result removed": {
			oldHook: map[string]any{"default_result": "CONTINUE"},
			newHook: map[string]any{"default_result": ""},
		},
		"hook is new": {
			newHook: map[string]any{metadata: "payload"},
		},
	}

	for name, testCase := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			if got := initialLifecycleHookNeedsReplacement(testCase.oldHook, testCase.newHook); got != testCase.want {
				t.Errorf("got %t, want %t", got, testCase.want)
			}
		})
	}
}

func TestInitialLifecycleHooksByName(t *testing.T) {
	t.Parallel()

	hooks := initialLifecycleHooksByName([]any{
		map[string]any{names.AttrName: "first", "notification_metadata": "payload"},
		nil,
		map[string]any{names.AttrName: ""},
		map[string]any{names.AttrName: "second"},
	})

	if len(hooks) != 2 {
		t.Fatalf("got %d hooks, want 2", len(hooks))
	}

	if got := hooks["first"]["notification_metadata"]; got != "payload" {
		t.Errorf(`hooks["first"]["notification_metadata"] = %v, want "payload"`, got)
	}
}
