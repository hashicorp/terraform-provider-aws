// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestLaunchTemplateStateUpgradeV0(t *testing.T) {
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
			name: "no elastic_gpu_specifications or elastic_inference_accelerator",
			rawState: map[string]any{
				names.AttrName: "test",
			},
			expected: map[string]any{
				names.AttrName: "test",
			},
		},
		{
			name: "with empty elastic_gpu_specifications",
			rawState: map[string]any{
				names.AttrName:                 "test",
				"elastic_gpu_specifications.#": "0",
			},
			expected: map[string]any{
				names.AttrName: "test",
			},
		},
		{
			name: "with empty elastic_inference_accelerator",
			rawState: map[string]any{
				names.AttrName:                    "test",
				"elastic_inference_accelerator.#": "0",
			},
			expected: map[string]any{
				names.AttrName: "test",
			},
		},
		{
			name: "with elastic_gpu_specifications and elastic_inference_accelerator",
			rawState: map[string]any{
				names.AttrName:                         "test",
				"elastic_gpu_specifications.#":         "1",
				"elastic_gpu_specifications.0.type":    "test1",
				"elastic_inference_accelerator.#":      "2",
				"elastic_inference_accelerator.0.type": "test2",
				"elastic_inference_accelerator.1.type": "test3",
			},
			expected: map[string]any{
				names.AttrName: "test",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfec2.LaunchTemplateStateUpgradeV0(t.Context(), tt.rawState, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
