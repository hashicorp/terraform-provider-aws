// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/names"
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

// TestInstanceEngineVersionDiffSuppress drives the real aws_db_instance schema
// diff (prior state + new config, no AWS) to prove the engine_version
// DiffSuppressFunc behaves correctly, including that d.Get observes the new
// (configured) auto_minor_version_upgrade within the same plan while
// engine_version_actual still reflects the prior/running version.
// Regression coverage for https://github.com/hashicorp/terraform-provider-aws/issues/39579.
func TestInstanceEngineVersionDiffSuppress(t *testing.T) {
	t.Parallel()

	type testCase struct {
		stateEngineVersion       string
		stateEngineVersionActual string
		stateAutoMinor           string
		configEngineVersion      string
		configAutoMinor          bool
		wantEngineVersionDiff    bool
		wantAutoMinorDiff        bool
	}
	testCases := map[string]testCase{
		// The critical case: engine_version relaxed to a major-only prefix and
		// auto_minor_version_upgrade flipped false -> true in the same plan. The
		// engine must not be downgraded; only auto_minor_version_upgrade changes.
		"same-plan false->true relaxes to major prefix": {
			stateEngineVersion:       "14.22",
			stateEngineVersionActual: "14.22",
			stateAutoMinor:           "false",
			configEngineVersion:      "14",
			configAutoMinor:          true,
			wantEngineVersionDiff:    false,
			wantAutoMinorDiff:        true,
		},
		// Automatic minor upgrades enabled and the running version is already
		// newer than the configured value (AWS advanced it out-of-band). No diff.
		"auto minor enabled, actual ahead of config": {
			stateEngineVersion:       "14.22",
			stateEngineVersionActual: "14.25",
			stateAutoMinor:           "true",
			configEngineVersion:      "14.22",
			configAutoMinor:          true,
			wantEngineVersionDiff:    false,
			wantAutoMinorDiff:        false,
		},
		// Automatic minor upgrades disabled: engine_version is an explicit pin, so
		// a lower configured version must still produce a (downgrade) diff.
		"explicit pin downgrade preserved": {
			stateEngineVersion:       "14.23",
			stateEngineVersionActual: "14.23",
			stateAutoMinor:           "false",
			configEngineVersion:      "14.22",
			configAutoMinor:          false,
			wantEngineVersionDiff:    true,
			wantAutoMinorDiff:        false,
		},
		// Intentional forward minor upgrade must always be preserved.
		"intentional forward upgrade preserved": {
			stateEngineVersion:       "14.22",
			stateEngineVersionActual: "14.22",
			stateAutoMinor:           "true",
			configEngineVersion:      "14.23",
			configAutoMinor:          true,
			wantEngineVersionDiff:    true,
			wantAutoMinorDiff:        false,
		},
		// Disabling automatic minor upgrades while leaving a major-only prefix
		// configured re-asserts Terraform ownership, so the prefix ("14") diffs
		// against the running minor ("14.22"). This surfaces a downgrade the user
		// must resolve by pinning engine_version to the exact running version; see
		// the engine_version documentation note.
		"auto minor disabled with prefix surfaces downgrade": {
			stateEngineVersion:       "14.22",
			stateEngineVersionActual: "14.22",
			stateAutoMinor:           "true",
			configEngineVersion:      "14",
			configAutoMinor:          false,
			wantEngineVersionDiff:    true,
			wantAutoMinorDiff:        true,
		},
	}

	r := resourceInstance()

	for name, test := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			state := &terraform.InstanceState{
				ID: "test",
				Attributes: map[string]string{
					names.AttrEngineVersion:           test.stateEngineVersion,
					"engine_version_actual":           test.stateEngineVersionActual,
					names.AttrAutoMinorVersionUpgrade: test.stateAutoMinor,
				},
			}
			config := terraform.NewResourceConfigRaw(map[string]any{
				names.AttrEngineVersion:           test.configEngineVersion,
				names.AttrAutoMinorVersionUpgrade: test.configAutoMinor,
			})

			diff, err := r.Diff(context.Background(), state, config, nil)
			if err != nil {
				t.Fatalf("computing diff: %s", err)
			}

			if got := hasAttrChange(diff, names.AttrEngineVersion); got != test.wantEngineVersionDiff {
				t.Errorf("engine_version diff = %t, want %t (attr=%#v)", got, test.wantEngineVersionDiff, attrDiff(diff, names.AttrEngineVersion))
			}
			if got := hasAttrChange(diff, names.AttrAutoMinorVersionUpgrade); got != test.wantAutoMinorDiff {
				t.Errorf("auto_minor_version_upgrade diff = %t, want %t", got, test.wantAutoMinorDiff)
			}
		})
	}
}

func attrDiff(diff *terraform.InstanceDiff, key string) *terraform.ResourceAttrDiff {
	if diff == nil {
		return nil
	}
	return diff.Attributes[key]
}

// hasAttrChange reports whether the diff contains a real (non-suppressed,
// non-NoOp) change for the given attribute.
func hasAttrChange(diff *terraform.InstanceDiff, key string) bool {
	a := attrDiff(diff, key)
	return a != nil && a.Old != a.New
}
