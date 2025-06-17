// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package batch_test

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	tfbatch "github.com/hashicorp/terraform-provider-aws/internal/service/batch"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestComputeEnvironmentStateUpgradeV0(t *testing.T) {
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
			name: "only compute_environment_name",
			rawState: map[string]any{
				"compute_environment_name": "test-environment",
			},
			expected: map[string]any{
				names.AttrName: "test-environment",
			},
		},
		{
			name: "only compute_environment_name_prefix",
			rawState: map[string]any{
				"compute_environment_name_prefix": "test-prefix",
			},
			expected: map[string]any{
				names.AttrNamePrefix: "test-prefix",
			},
		},
		{
			name: "both compute_environment_name and compute_environment_name_prefix",
			rawState: map[string]any{
				"compute_environment_name":        "test-environment",
				"compute_environment_name_prefix": "test-prefix",
			},
			expected: map[string]any{
				names.AttrName:       "test-environment",
				names.AttrNamePrefix: "test-prefix",
			},
		},
		{
			name: "unrelated keys in rawState",
			rawState: map[string]any{
				"compute_environment_name":        "test-environment",
				"compute_environment_name_prefix": "test-prefix",
				"other_key":                       "other-value",
			},
			expected: map[string]any{
				names.AttrName:       "test-environment",
				names.AttrNamePrefix: "test-prefix",
				"other_key":          "other-value",
			},
		},
		{
			name: "no compute_environment_name or compute_environment_name_prefix",
			rawState: map[string]any{
				"other_key": "other-value",
			},
			expected: map[string]any{
				"other_key": "other-value",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfbatch.ComputeEnvironmentStateUpgradeV0(context.Background(), tt.rawState, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
