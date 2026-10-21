// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"strconv"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func TestInstanceReplicateSourceDBSuppressDiff(t *testing.T) {
	t.Parallel()

	type testCase struct {
		old, new string
		expected bool
	}
	testCases := map[string]testCase{
		"no values": {
			old:      "",
			new:      "",
			expected: false,
		},

		"old ARN same identifier": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test", //lintignore:AWSAT003,AWSAT005
			new:      "test",
			expected: true,
		},

		"old ARN different identifier": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			new:      "test2",
			expected: false,
		},

		"new ARN same identifier": {
			old:      "test",
			new:      "arn:aws:rds:us-west-2:123456789012:db:test", //lintignore:AWSAT003,AWSAT005
			expected: true,
		},

		"new ARN different identifier": {
			old:      "test2",
			new:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			expected: false,
		},

		"both ARN": {
			old:      "arn:aws:rds:us-west-2:123456789012:db:test1", //lintignore:AWSAT003,AWSAT005
			new:      "arn:aws:rds:us-west-2:123456789012:db:test2", //lintignore:AWSAT003,AWSAT005
			expected: false,
		},

		"neither ARN": {
			old:      "test1",
			new:      "test2",
			expected: false,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			v := instanceReplicateSourceDBSuppressDiff("", test.old, test.new, nil)
			if e, a := test.expected, v; e != a {
				t.Errorf("unexpected result: expected %t, got %t", e, a)
			}
		})
	}
}

func TestInstanceMultiTenantForceNew(t *testing.T) {
	t.Parallel()

	// Legacy SDK representation of an unknown value (hcl2shim.UnknownVariableValue).
	const unknown = "74D93920-ED26-11E3-AC10-0800200C9A66"

	type testCase struct {
		state           *bool // multi_tenant in prior state; nil means absent (e.g. state from an earlier provider version)
		config          any   // multi_tenant in configuration; nil means not set
		wantDiff        bool
		wantRequiresNew bool
	}
	testCases := map[string]testCase{
		"absent from state, not set": {
			config: nil,
		},
		"absent from state, single-tenant to multi-tenant": {
			config:   true,
			wantDiff: true,
		},
		"single-tenant, not set": {
			state:  aws.Bool(false),
			config: nil,
		},
		"multi-tenant, not set": {
			state:  aws.Bool(true),
			config: nil,
		},
		"multi-tenant, unchanged": {
			state:  aws.Bool(true),
			config: true,
		},
		"single-tenant to multi-tenant": {
			state:    aws.Bool(false),
			config:   true,
			wantDiff: true,
		},
		"multi-tenant to single-tenant": {
			state:           aws.Bool(true),
			config:          false,
			wantDiff:        true,
			wantRequiresNew: true,
		},
		"multi-tenant to unknown": {
			state:    aws.Bool(true),
			config:   unknown,
			wantDiff: true,
		},
	}

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			ctx := t.Context()

			raw := map[string]any{
				"instance_class": "db.t3.small",
			}
			if test.config != nil {
				raw["multi_tenant"] = test.config
			}

			state := &terraform.InstanceState{
				ID: "db-ABCDEFGHIJKLMNOPQRSTUVWXYZ",
				Attributes: map[string]string{
					"instance_class": "db.t3.small",
				},
			}
			if test.state != nil {
				state.Attributes["multi_tenant"] = strconv.FormatBool(*test.state)
			}

			diff, err := resourceInstance().Diff(ctx, state, terraform.NewResourceConfigRaw(raw), nil)
			if err != nil {
				t.Fatalf("unexpected error: %s", err)
			}

			var attr *terraform.ResourceAttrDiff
			if diff != nil {
				attr = diff.Attributes["multi_tenant"]
			}

			if got := attr != nil; got != test.wantDiff {
				t.Fatalf("multi_tenant diff: expected %t, got %t (%#v)", test.wantDiff, got, attr)
			}
			if got := diff.RequiresNew(); got != test.wantRequiresNew {
				t.Errorf("requires new: expected %t, got %t", test.wantRequiresNew, got)
			}
		})
	}
}
