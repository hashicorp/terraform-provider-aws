// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
)

func TestInstanceMigrateState(t *testing.T) {
	t.Parallel()

	cases := map[string]struct {
		StateVersion int
		Attributes   map[string]string
		Expected     map[string]string
		Meta         any
	}{
		"v0.3.6 and earlier": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"block_device.#": "2",
				"block_device.3851383343.delete_on_termination": acctest.CtTrue,
				"block_device.3851383343.device_name":           "/dev/sdx",
				"block_device.3851383343.encrypted":             acctest.CtFalse,
				"block_device.3851383343.snapshot_id":           "",
				"block_device.3851383343.virtual_name":          "",
				"block_device.3851383343.volume_size":           "5",
				"block_device.3851383343.volume_type":           "standard",
				// Ephemeral
				"block_device.3101711606.delete_on_termination": acctest.CtFalse,
				"block_device.3101711606.device_name":           "/dev/sdy",
				"block_device.3101711606.encrypted":             acctest.CtFalse,
				"block_device.3101711606.snapshot_id":           "",
				"block_device.3101711606.virtual_name":          "ephemeral0",
				"block_device.3101711606.volume_size":           "",
				"block_device.3101711606.volume_type":           "",
				// Root
				"block_device.56575650.delete_on_termination": acctest.CtTrue,
				"block_device.56575650.device_name":           "/dev/sda1",
				"block_device.56575650.encrypted":             acctest.CtFalse,
				"block_device.56575650.snapshot_id":           "",
				"block_device.56575650.volume_size":           "10",
				"block_device.56575650.volume_type":           "standard",
			},
			Expected: map[string]string{
				"ebs_block_device.#": "1",
				"ebs_block_device.3851383343.delete_on_termination":  acctest.CtTrue,
				"ebs_block_device.3851383343.device_name":            "/dev/sdx",
				"ebs_block_device.3851383343.encrypted":              acctest.CtFalse,
				"ebs_block_device.3851383343.snapshot_id":            "",
				"ebs_block_device.3851383343.volume_size":            "5",
				"ebs_block_device.3851383343.volume_type":            "standard",
				"ephemeral_block_device.#":                           "1",
				"ephemeral_block_device.2458403513.device_name":      "/dev/sdy",
				"ephemeral_block_device.2458403513.virtual_name":     "ephemeral0",
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "standard",
			},
		},
		"v0.3.7": {
			StateVersion: 0,
			Attributes: map[string]string{
				// EBS
				"block_device.#": "2",
				"block_device.3851383343.delete_on_termination": acctest.CtTrue,
				"block_device.3851383343.device_name":           "/dev/sdx",
				"block_device.3851383343.encrypted":             acctest.CtFalse,
				"block_device.3851383343.snapshot_id":           "",
				"block_device.3851383343.virtual_name":          "",
				"block_device.3851383343.volume_size":           "5",
				"block_device.3851383343.volume_type":           "standard",
				"block_device.3851383343.iops":                  "",
				// Ephemeral
				"block_device.3101711606.delete_on_termination": acctest.CtFalse,
				"block_device.3101711606.device_name":           "/dev/sdy",
				"block_device.3101711606.encrypted":             acctest.CtFalse,
				"block_device.3101711606.snapshot_id":           "",
				"block_device.3101711606.virtual_name":          "ephemeral0",
				"block_device.3101711606.volume_size":           "",
				"block_device.3101711606.volume_type":           "",
				"block_device.3101711606.iops":                  "",
				// Root
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "io1",
				"root_block_device.3018388612.iops":                  "1000",
			},
			Expected: map[string]string{
				"ebs_block_device.#": "1",
				"ebs_block_device.3851383343.delete_on_termination":  acctest.CtTrue,
				"ebs_block_device.3851383343.device_name":            "/dev/sdx",
				"ebs_block_device.3851383343.encrypted":              acctest.CtFalse,
				"ebs_block_device.3851383343.snapshot_id":            "",
				"ebs_block_device.3851383343.volume_size":            "5",
				"ebs_block_device.3851383343.volume_type":            "standard",
				"ephemeral_block_device.#":                           "1",
				"ephemeral_block_device.2458403513.device_name":      "/dev/sdy",
				"ephemeral_block_device.2458403513.virtual_name":     "ephemeral0",
				"root_block_device.#":                                "1",
				"root_block_device.3018388612.delete_on_termination": acctest.CtTrue,
				"root_block_device.3018388612.device_name":           "/dev/sda1",
				"root_block_device.3018388612.snapshot_id":           "",
				"root_block_device.3018388612.volume_size":           "10",
				"root_block_device.3018388612.volume_type":           "io1",
				"root_block_device.3018388612.iops":                  "1000",
			},
		},
	}

	for tn, tc := range cases {
		is := &terraform.InstanceState{
			ID:         "i-abc123",
			Attributes: tc.Attributes,
		}
		is, err := tfec2.InstanceMigrateState(
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

func TestInstanceMigrateState_empty(t *testing.T) {
	t.Parallel()

	var is *terraform.InstanceState
	var meta any

	// should handle nil
	is, err := tfec2.InstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
	if is != nil {
		t.Fatalf("expected nil instancestate, got: %#v", is)
	}

	// should handle non-nil but empty
	is = &terraform.InstanceState{}
	_, err = tfec2.InstanceMigrateState(0, is, meta)

	if err != nil {
		t.Fatalf("err: %#v", err)
	}
}

func TestInstanceStateUpgradeV1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rawState map[string]any
		expected map[string]any
	}{
		{
			name:     "empty rawState",
			rawState: nil,
			expected: map[string]any{},
		},
		{
			name: "no cpu_options, only cpu_core_count",
			rawState: map[string]any{
				"cpu_core_count": 4,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count": 4,
					},
				},
			},
		},
		{
			name: "no cpu_options, only cpu_threads_per_core",
			rawState: map[string]any{
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "no cpu_options, both cpu_core_count and cpu_threads_per_core",
			rawState: map[string]any{
				"cpu_core_count":       4,
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       4,
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "existing cpu_options with core_count and threads_per_core",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
				"cpu_core_count":       4,
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
			},
		},
		{
			name: "existing cpu_options with only core_count",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count": 8,
					},
				},
				"cpu_threads_per_core": 2,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 2,
					},
				},
			},
		},
		{
			name: "existing cpu_options with only threads_per_core",
			rawState: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"threads_per_core": 4,
					},
				},
				"cpu_core_count": 8,
			},
			expected: map[string]any{
				"cpu_options": []any{
					map[string]any{
						"core_count":       8,
						"threads_per_core": 4,
					},
				},
			},
		},
		{
			name: "no cpu_options and no cpu_core_count or cpu_threads_per_core",
			rawState: map[string]any{
				"some_other_key": "some_value",
			},
			expected: map[string]any{
				"some_other_key": "some_value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfec2.InstanceStateUpgradeV1(context.Background(), tt.rawState, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
