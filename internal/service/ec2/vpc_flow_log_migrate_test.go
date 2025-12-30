// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"testing"

	"github.com/google/go-cmp/cmp"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestFlowLogStateUpgradeV0(t *testing.T) {
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
			name: "no log_group_name",
			rawState: map[string]any{
				names.AttrSubnetID: "sn-12345678",
			},
			expected: map[string]any{
				names.AttrSubnetID: "sn-12345678",
			},
		},
		{
			name: "with log_group_name",
			rawState: map[string]any{
				names.AttrLogGroupName: "log-group-name",
				names.AttrSubnetID:     "sn-12345678",
			},
			expected: map[string]any{
				names.AttrSubnetID: "sn-12345678",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result, err := tfec2.FlowLogStateUpgradeV0(t.Context(), tt.rawState, nil)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if diff := cmp.Diff(tt.expected, result); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
