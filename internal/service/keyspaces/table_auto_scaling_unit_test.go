// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package keyspaces_test

import (
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/keyspaces/types"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	tfkeyspaces "github.com/hashicorp/terraform-provider-aws/internal/service/keyspaces"
)

var ignoreUnexported = cmpopts.IgnoreUnexported(
	types.AutoScalingSpecification{},
	types.AutoScalingSettings{},
	types.AutoScalingPolicy{},
	types.TargetTrackingScalingPolicyConfiguration{},
)

func TestExpandTargetTrackingScalingPolicyConfiguration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfMap    map[string]any
		expected *types.TargetTrackingScalingPolicyConfiguration
	}{
		"nil map": {
			tfMap:    nil,
			expected: nil,
		},
		"zero-value bools and cooldowns are preserved, not dropped": {
			tfMap: map[string]any{
				"disable_scale_in":   false,
				"scale_in_cooldown":  0,
				"scale_out_cooldown": 0,
				"target_value":       70.0,
			},
			expected: &types.TargetTrackingScalingPolicyConfiguration{
				DisableScaleIn:   false,
				ScaleInCooldown:  0,
				ScaleOutCooldown: 0,
				TargetValue:      70.0,
			},
		},
		"non-zero-value bools and cooldowns": {
			tfMap: map[string]any{
				"disable_scale_in":   true,
				"scale_in_cooldown":  60,
				"scale_out_cooldown": 120,
				"target_value":       55.5,
			},
			expected: &types.TargetTrackingScalingPolicyConfiguration{
				DisableScaleIn:   true,
				ScaleInCooldown:  60,
				ScaleOutCooldown: 120,
				TargetValue:      55.5,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.ExpandTargetTrackingScalingPolicyConfiguration(tc.tfMap)

			if diff := cmp.Diff(tc.expected, got, ignoreUnexported); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlattenTargetTrackingScalingPolicyConfiguration(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		apiObject *types.TargetTrackingScalingPolicyConfiguration
		expected  map[string]any
	}{
		"nil": {
			apiObject: nil,
			expected:  nil,
		},
		"zero-value bools and cooldowns surface, not omitted": {
			apiObject: &types.TargetTrackingScalingPolicyConfiguration{
				DisableScaleIn:   false,
				ScaleInCooldown:  0,
				ScaleOutCooldown: 0,
				TargetValue:      70.0,
			},
			expected: map[string]any{
				"disable_scale_in":   false,
				"scale_in_cooldown":  int32(0),
				"scale_out_cooldown": int32(0),
				"target_value":       70.0,
			},
		},
		"non-zero-value bools and cooldowns": {
			apiObject: &types.TargetTrackingScalingPolicyConfiguration{
				DisableScaleIn:   true,
				ScaleInCooldown:  60,
				ScaleOutCooldown: 120,
				TargetValue:      55.5,
			},
			expected: map[string]any{
				"disable_scale_in":   true,
				"scale_in_cooldown":  int32(60),
				"scale_out_cooldown": int32(120),
				"target_value":       55.5,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.FlattenTargetTrackingScalingPolicyConfiguration(tc.apiObject)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExpandAutoScalingSettings(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		tfMap    map[string]any
		expected *types.AutoScalingSettings
	}{
		"nil map": {
			tfMap:    nil,
			expected: nil,
		},
		"explicit false auto_scaling_disabled is preserved, with units and nested policy expanded": {
			tfMap: map[string]any{
				"auto_scaling_disabled": false,
				"minimum_units":         5,
				"maximum_units":         10,
				"target_tracking_scaling_policy_configuration": []any{
					map[string]any{
						"disable_scale_in":   false,
						"scale_in_cooldown":  0,
						"scale_out_cooldown": 0,
						"target_value":       70.0,
					},
				},
			},
			expected: &types.AutoScalingSettings{
				AutoScalingDisabled: false,
				MinimumUnits:        aws.Int64(5),
				MaximumUnits:        aws.Int64(10),
				ScalingPolicy: &types.AutoScalingPolicy{
					TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
						TargetValue: 70.0,
					},
				},
			},
		},
		"auto_scaling_disabled true omits minimum/maximum units and scaling policy": {
			// AWS rejects the update with "When disabled, auto scaling settings should not be
			// provided" if these are also sent alongside AutoScalingDisabled: true.
			tfMap: map[string]any{
				"auto_scaling_disabled": true,
				"minimum_units":         5,
				"maximum_units":         10,
				"target_tracking_scaling_policy_configuration": []any{
					map[string]any{"target_value": 70.0},
				},
			},
			expected: &types.AutoScalingSettings{
				AutoScalingDisabled: true,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.ExpandAutoScalingSettings(tc.tfMap)

			if diff := cmp.Diff(tc.expected, got, ignoreUnexported); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlattenAutoScalingSettings(t *testing.T) {
	t.Parallel()

	testCases := map[string]struct {
		apiObject *types.AutoScalingSettings
		expected  map[string]any
	}{
		"nil": {
			apiObject: nil,
			expected:  nil,
		},
		"false auto_scaling_disabled and units surface; nil scaling policy is omitted": {
			apiObject: &types.AutoScalingSettings{
				AutoScalingDisabled: false,
				MinimumUnits:        aws.Int64(5),
				MaximumUnits:        aws.Int64(10),
			},
			expected: map[string]any{
				"auto_scaling_disabled": false,
				"minimum_units":         int64(5),
				"maximum_units":         int64(10),
			},
		},
		"nil MinimumUnits/MaximumUnits pointers are omitted from the map": {
			apiObject: &types.AutoScalingSettings{
				AutoScalingDisabled: true,
			},
			expected: map[string]any{
				"auto_scaling_disabled": true,
			},
		},
		"full round trip includes the nested scaling policy": {
			apiObject: &types.AutoScalingSettings{
				AutoScalingDisabled: false,
				MinimumUnits:        aws.Int64(5),
				MaximumUnits:        aws.Int64(10),
				ScalingPolicy: &types.AutoScalingPolicy{
					TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
						TargetValue: 70.0,
					},
				},
			},
			expected: map[string]any{
				"auto_scaling_disabled": false,
				"minimum_units":         int64(5),
				"maximum_units":         int64(10),
				"target_tracking_scaling_policy_configuration": []any{
					map[string]any{
						"disable_scale_in":   false,
						"scale_in_cooldown":  int32(0),
						"scale_out_cooldown": int32(0),
						"target_value":       70.0,
					},
				},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.FlattenAutoScalingSettings(tc.apiObject)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExpandAutoScalingSpecification(t *testing.T) {
	t.Parallel()

	settingsTfMap := map[string]any{
		"minimum_units": 5,
		"maximum_units": 10,
		"target_tracking_scaling_policy_configuration": []any{
			map[string]any{"target_value": 70.0},
		},
	}
	expandedSettings := &types.AutoScalingSettings{
		MinimumUnits: aws.Int64(5),
		MaximumUnits: aws.Int64(10),
		ScalingPolicy: &types.AutoScalingPolicy{
			TargetTrackingScalingPolicyConfiguration: &types.TargetTrackingScalingPolicyConfiguration{
				TargetValue: 70.0,
			},
		},
	}

	testCases := map[string]struct {
		tfMap    map[string]any
		expected *types.AutoScalingSpecification
	}{
		"nil map": {
			tfMap:    nil,
			expected: nil,
		},
		"neither read nor write set": {
			tfMap:    map[string]any{},
			expected: &types.AutoScalingSpecification{},
		},
		"only read_capacity_auto_scaling set": {
			tfMap: map[string]any{
				"read_capacity_auto_scaling": []any{settingsTfMap},
			},
			expected: &types.AutoScalingSpecification{
				ReadCapacityAutoScaling: expandedSettings,
			},
		},
		"only write_capacity_auto_scaling set": {
			tfMap: map[string]any{
				"write_capacity_auto_scaling": []any{settingsTfMap},
			},
			expected: &types.AutoScalingSpecification{
				WriteCapacityAutoScaling: expandedSettings,
			},
		},
		"both read and write set": {
			tfMap: map[string]any{
				"read_capacity_auto_scaling":  []any{settingsTfMap},
				"write_capacity_auto_scaling": []any{settingsTfMap},
			},
			expected: &types.AutoScalingSpecification{
				ReadCapacityAutoScaling:  expandedSettings,
				WriteCapacityAutoScaling: expandedSettings,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.ExpandAutoScalingSpecification(tc.tfMap)

			if diff := cmp.Diff(tc.expected, got, ignoreUnexported); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestFlattenAutoScalingSpecification(t *testing.T) {
	t.Parallel()

	settings := &types.AutoScalingSettings{
		MinimumUnits: aws.Int64(5),
		MaximumUnits: aws.Int64(10),
	}
	flattenedSettings := map[string]any{
		"auto_scaling_disabled": false,
		"minimum_units":         int64(5),
		"maximum_units":         int64(10),
	}

	testCases := map[string]struct {
		apiObject *types.AutoScalingSpecification
		expected  map[string]any
	}{
		"nil": {
			apiObject: nil,
			expected:  nil,
		},
		"only ReadCapacityAutoScaling set": {
			apiObject: &types.AutoScalingSpecification{
				ReadCapacityAutoScaling: settings,
			},
			expected: map[string]any{
				"read_capacity_auto_scaling": []any{flattenedSettings},
			},
		},
		"only WriteCapacityAutoScaling set": {
			apiObject: &types.AutoScalingSpecification{
				WriteCapacityAutoScaling: settings,
			},
			expected: map[string]any{
				"write_capacity_auto_scaling": []any{flattenedSettings},
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.FlattenAutoScalingSpecification(tc.apiObject)

			if diff := cmp.Diff(tc.expected, got); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}

func TestExpandAutoScalingSpecificationDisabled(t *testing.T) {
	t.Parallel()

	// AWS rejects the update with "When disabled, auto scaling settings should not be
	// provided" if minimum_units/maximum_units/scaling_policy are sent alongside
	// AutoScalingDisabled: true, so the prior values must NOT be preserved.
	settingsTfMap := map[string]any{
		"auto_scaling_disabled": false,
		"minimum_units":         5,
		"maximum_units":         10,
		"target_tracking_scaling_policy_configuration": []any{
			map[string]any{"target_value": 70.0},
		},
	}
	disabled := &types.AutoScalingSettings{AutoScalingDisabled: true}

	testCases := map[string]struct {
		tfMap    map[string]any
		expected *types.AutoScalingSpecification
	}{
		"nil map does not panic and returns nil": {
			tfMap:    nil,
			expected: nil,
		},
		"forces AutoScalingDisabled true on both settings and omits prior values": {
			tfMap: map[string]any{
				"read_capacity_auto_scaling":  []any{settingsTfMap},
				"write_capacity_auto_scaling": []any{settingsTfMap},
			},
			expected: &types.AutoScalingSpecification{
				ReadCapacityAutoScaling:  disabled,
				WriteCapacityAutoScaling: disabled,
			},
		},
		"only read set does not populate write": {
			tfMap: map[string]any{
				"read_capacity_auto_scaling": []any{settingsTfMap},
			},
			expected: &types.AutoScalingSpecification{
				ReadCapacityAutoScaling: disabled,
			},
		},
	}

	for name, tc := range testCases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			got := tfkeyspaces.ExpandAutoScalingSpecificationDisabled(tc.tfMap)

			if diff := cmp.Diff(tc.expected, got, ignoreUnexported); diff != "" {
				t.Errorf("unexpected result (-want +got):\n%s", diff)
			}
		})
	}
}
