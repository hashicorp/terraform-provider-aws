// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"testing"

	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
)

func TestArePromptAttemptsEqual(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	tests := []struct {
		name string
		// plannedState represents the terraform object after a refresh (remote objects)
		plannedState fwtypes.SetNestedObjectValueOf[tflexv2models.PromptAttemptsSpecification]
		// incomingPlan represents the terraform object after a plan (plan to match config)
		incomingPlan fwtypes.SetNestedObjectValueOf[tflexv2models.PromptAttemptsSpecification]
		maxRetries   int64
		expected     bool
	}{
		{
			name: "new value without default",
			plannedState: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
			}),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
			}),
			maxRetries: 1,
			expected:   true,
		},
		{
			name: "new value with specified default",
			plannedState: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
			}),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry2"),
			}),
			maxRetries: 2,
			expected:   false,
		},
		{
			name: "new values of null",
			plannedState: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry2"),
			}),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfNull[tflexv2models.PromptAttemptsSpecification](ctx),
			maxRetries:   2,
			expected:     true,
		},
		{
			name:         "old values of null",
			plannedState: fwtypes.NewSetNestedObjectValueOfNull[tflexv2models.PromptAttemptsSpecification](ctx),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry2"),
			}),
			maxRetries: 2,
			expected:   false,
		},
		{
			name: "missing default values",
			plannedState: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry2"),
			}),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry2"),
			}),
			maxRetries: 2,
			expected:   true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			result, _ := tflexv2models.ArePromptAttemptsEqual(ctx, test.plannedState, test.incomingPlan, test.maxRetries)
			if result != test.expected {
				t.Errorf("expected %v, got %v", test.expected, result)
			}
		})
	}
}
