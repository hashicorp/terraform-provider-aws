// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"testing"

	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
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
			// represent no defaults being set in the configuration
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
			name: "configured value different from default",
			plannedState: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Initial"),
				tflexv2models.DefaultPromptAttemptsSpecification(ctx, "Retry1"),
			}),
			incomingPlan: fwtypes.NewSetNestedObjectValueOfSliceMust[tflexv2models.PromptAttemptsSpecification](ctx, []*tflexv2models.PromptAttemptsSpecification{
				{
					MapBlockKey: fwtypes.StringEnumValue(tflexv2models.PromptAttemptsType("Initial")),
					AllowedInputTypes: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tflexv2models.AllowedInputTypes{
						{
							AllowAudioInput: fwflex.BoolValueToFramework(ctx, true),
							AllowDTMFInput:  fwflex.BoolValueToFramework(ctx, true),
						},
					}),
					AllowInterrupt: fwflex.BoolValueToFramework(ctx, false),
					AudioAndDTMFInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tflexv2models.AudioAndDTMFInputSpecification{
						{
							StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 4000), //nolint:mnd
							AudioSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tflexv2models.AudioSpecification{
								{
									EndTimeoutMs: fwflex.Int64ValueToFramework(ctx, 640),   //nolint:mnd
									MaxLengthMs:  fwflex.Int64ValueToFramework(ctx, 15000), //nolint:mnd
								},
							}),
							DTMFSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tflexv2models.DTMFSpecification{
								{
									DeletionCharacter: fwflex.StringValueToFramework(ctx, "*"),
									EndCharacter:      fwflex.StringValueToFramework(ctx, "#"),
									EndTimeoutMs:      fwflex.Int64ValueToFramework(ctx, 5000), //nolint:mnd
									MaxLength:         fwflex.Int64ValueToFramework(ctx, 513),  //nolint:mnd
								},
							}),
						},
					}),
					TextInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*tflexv2models.TextInputSpecification{
						{
							StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 30000), //nolint:mnd
						},
					}),
				},
			}),
			maxRetries: 2,
			expected:   false,
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
