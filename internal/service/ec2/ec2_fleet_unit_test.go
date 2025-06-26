// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	awstypes "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

// Unit test for flattenSpotOptions function to verify instance_pools_to_use_count handling
func TestFlattenSpotOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    *awstypes.SpotOptions
		expected map[string]any
	}{
		{
			name: "lowestPrice with InstancePoolsToUseCount",
			input: &awstypes.SpotOptions{
				AllocationStrategy:      awstypes.SpotAllocationStrategyLowestPrice,
				InstancePoolsToUseCount: aws.Int32(2),
			},
			expected: map[string]any{
				"allocation_strategy":         awstypes.SpotAllocationStrategyLowestPrice,
				"instance_pools_to_use_count": int32(2),
			},
		},
		{
			name: "lowestPrice without InstancePoolsToUseCount",
			input: &awstypes.SpotOptions{
				AllocationStrategy: awstypes.SpotAllocationStrategyLowestPrice,
			},
			expected: map[string]any{
				"allocation_strategy": awstypes.SpotAllocationStrategyLowestPrice,
			},
		},
		{
			name: "price-capacity-optimized without InstancePoolsToUseCount",
			input: &awstypes.SpotOptions{
				AllocationStrategy: awstypes.SpotAllocationStrategyPriceCapacityOptimized,
			},
			expected: map[string]any{
				"allocation_strategy":         awstypes.SpotAllocationStrategyPriceCapacityOptimized,
				"instance_pools_to_use_count": 1, // Should default to 1
			},
		},
		{
			name: "diversified without InstancePoolsToUseCount",
			input: &awstypes.SpotOptions{
				AllocationStrategy: awstypes.SpotAllocationStrategyDiversified,
			},
			expected: map[string]any{
				"allocation_strategy":         awstypes.SpotAllocationStrategyDiversified,
				"instance_pools_to_use_count": 1, // Should default to 1
			},
		},
		{
			name: "capacity-optimized without InstancePoolsToUseCount",
			input: &awstypes.SpotOptions{
				AllocationStrategy: awstypes.SpotAllocationStrategyCapacityOptimized,
			},
			expected: map[string]any{
				"allocation_strategy":         awstypes.SpotAllocationStrategyCapacityOptimized,
				"instance_pools_to_use_count": 1, // Should default to 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenSpotOptions(tt.input)

			for key, expectedValue := range tt.expected {
				if actualValue, ok := result[key]; !ok {
					t.Errorf("Expected key %s not found in result", key)
				} else if actualValue != expectedValue {
					t.Errorf("For key %s, expected %v, got %v", key, expectedValue, actualValue)
				}
			}
		})
	}
}
