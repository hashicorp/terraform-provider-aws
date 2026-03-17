// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"encoding/json"
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

// TestFlowLogStateUpgradeV0_complexState simulates state as decoded from a JSON state file.
// This is to ensure nothing in a complex state prevents state upgrading.
func TestFlowLogStateUpgradeV0_complexState(t *testing.T) {
	t.Parallel()

	stateJSON := `{
		"arn": "arn:aws:ec2:us-east-1:123456789012:vpc-flow-log/fl-12345678",
		"deliver_cross_account_role": "",
		"destination_options": [],
		"eni_id": "eni-12345678",
		"iam_role_arn": "arn:aws:iam::123456789012:role/flowlogs",
		"id": "fl-12345678",
		"log_destination": "arn:aws:logs:us-east-1:123456789012:log-group:/my/log-group",
		"log_destination_type": "cloud-watch-logs",
		"log_format": "${version} ${account-id}",
		"log_group_name": "/my/log-group",
		"max_aggregation_interval": 600,
		"subnet_id": null,
		"tags": {},
		"tags_all": {},
		"traffic_type": "ALL",
		"transit_gateway_attachment_id": null,
		"transit_gateway_id": null,
		"vpc_id": null
	}` //lintignore:AWSAT003,AWSAT005

	var rawState map[string]any
	if err := json.Unmarshal([]byte(stateJSON), &rawState); err != nil {
		t.Fatalf("failed to unmarshal state JSON: %v", err)
	}

	result, err := tfec2.FlowLogStateUpgradeV0(t.Context(), rawState, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, ok := result[names.AttrLogGroupName]; ok {
		t.Errorf("expected log_group_name to be removed, but it is still present")
	}
	//lintignore:AWSAT003,AWSAT005
	if result["log_destination"] != "arn:aws:logs:us-east-1:123456789012:log-group:/my/log-group" {
		t.Errorf("expected log_destination to be preserved, got: %v", result["log_destination"])
	}
}
