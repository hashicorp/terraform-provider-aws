// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"fmt"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// confirmationSettingsEqualityFunc checks semantic equality.
// This specifically targets default values on PromptAttemptsSpecification
func confirmationSettingsEqualityFunc(ctx context.Context, oldValue, newValue fwtypes.NestedCollectionValue[IntentConfirmationSetting]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldConfirmationSettings, di := oldValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	newConfirmationSettings, di := newValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	if oldConfirmationSettings != nil && newConfirmationSettings != nil {
		if !oldConfirmationSettings.CodeHook.Equal(newConfirmationSettings.CodeHook) ||
			!oldConfirmationSettings.ConfirmationConditional.Equal(newConfirmationSettings.ConfirmationConditional) ||
			!oldConfirmationSettings.ConfirmationResponse.Equal(newConfirmationSettings.ConfirmationResponse) ||
			!oldConfirmationSettings.ConfirmationNextStep.Equal(newConfirmationSettings.ConfirmationNextStep) ||
			!oldConfirmationSettings.DeclinationConditional.Equal(newConfirmationSettings.DeclinationConditional) ||
			!oldConfirmationSettings.DeclinationNextStep.Equal(newConfirmationSettings.DeclinationNextStep) ||
			!oldConfirmationSettings.DeclinationResponse.Equal(newConfirmationSettings.DeclinationResponse) ||
			!oldConfirmationSettings.ElicitationCodeHook.Equal(newConfirmationSettings.ElicitationCodeHook) ||
			!oldConfirmationSettings.FailureConditional.Equal(newConfirmationSettings.FailureConditional) ||
			!oldConfirmationSettings.FailureNextStep.Equal(newConfirmationSettings.FailureNextStep) ||
			!oldConfirmationSettings.FailureResponse.Equal(newConfirmationSettings.FailureResponse) {
			return false, diags
		}

		oldPromptSpec, di := oldConfirmationSettings.PromptSpecification.ToPtr(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		newPromptSpec, di := newConfirmationSettings.PromptSpecification.ToPtr(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		if !oldPromptSpec.AllowInterrupt.Equal(newPromptSpec.AllowInterrupt) ||
			!oldPromptSpec.MaxRetries.Equal(newPromptSpec.MaxRetries) ||
			!oldPromptSpec.MessageGroup.Equal(newPromptSpec.MessageGroup) ||
			!oldPromptSpec.MessageSelectionStrategy.Equal(newPromptSpec.MessageSelectionStrategy) {
			return false, diags
		}

		return arePromptAttemptsEqual(ctx, oldPromptSpec.PromptAttemptsSpecification, newPromptSpec.PromptAttemptsSpecification, oldPromptSpec.MaxRetries.ValueInt64())
	}

	return false, diags
}

// arePromptAttemptsEqual compares two PromptAttemptsSpecification fields for equality
// treating them as maps with map_block_key as the key
func arePromptAttemptsEqual(ctx context.Context, oldAttempts, newAttempts fwtypes.SetNestedObjectValueOf[PromptAttemptsSpecification], maxRetries int64) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics
	// If both are null or unknown, they're equal
	if oldAttempts.IsNull() && newAttempts.IsNull() {
		return true, diags
	}
	if oldAttempts.IsUnknown() && newAttempts.IsUnknown() {
		return true, diags
	}

	if !oldAttempts.Equal(newAttempts) {
		// Convert to slices for comparison
		oldPromptAttemptSpecification, di := oldAttempts.ToSlice(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		newPromptAttemptSpecification, di := newAttempts.ToSlice(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		pasExists := promptAttemptsSpecificationDefaults(ctx, maxRetries)
		var hasDefaults, areEqual bool
		for _, value := range oldPromptAttemptSpecification {
			key := value.MapBlockKey.ValueString()
			index := slices.IndexFunc(newPromptAttemptSpecification, func(item *PromptAttemptsSpecification) bool {
				return item.MapBlockKey.ValueString() == key
			})

			if index != -1 {
				areEqual = arePromptAttemptValuesEqual(*newPromptAttemptSpecification[index], *value)
			}

			_, ok := pasExists(key)
			hasDefaults = ok
		}

		return (hasDefaults && areEqual) || len(newPromptAttemptSpecification) == 0, diags
	}

	return false, diags
}

// arePromptAttemptValuesEqual compares two PromptAttemptsSpecification items for equality
// ignoring the map_block_key field since it's used as the map key
func arePromptAttemptValuesEqual(oldItem, newItem PromptAttemptsSpecification) bool {
	return oldItem.AllowInterrupt.Equal(newItem.AllowInterrupt) &&
		oldItem.AllowedInputTypes.Equal(newItem.AllowedInputTypes) &&
		oldItem.AudioAndDTMFInputSpecification.Equal(newItem.AudioAndDTMFInputSpecification) &&
		oldItem.TextInputSpecification.Equal(newItem.TextInputSpecification)
}

func promptAttemptsSpecificationDefaults(ctx context.Context, maxRetries int64) func(string) (*PromptAttemptsSpecification, bool) {
	defaults := map[string]*PromptAttemptsSpecification{
		"Initial": defaultPromptAttemptsSpecification(ctx, "Initial"),
	}

	for i := range maxRetries {
		k := fmt.Sprintf("Retry%d", i+1)
		defaults[k] = defaultPromptAttemptsSpecification(ctx, PromptAttemptsType(k))
	}

	return func(key string) (*PromptAttemptsSpecification, bool) {
		if val, ok := defaults[key]; ok {
			return val, true
		}

		return nil, false
	}
}

func defaultPromptAttemptsSpecification(ctx context.Context, mapBlockKey PromptAttemptsType) *PromptAttemptsSpecification {
	return &PromptAttemptsSpecification{
		MapBlockKey: fwtypes.StringEnumValue(mapBlockKey),
		AllowedInputTypes: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AllowedInputTypes{
			{
				AllowAudioInput: fwflex.BoolValueToFramework(ctx, true),
				AllowDTMFInput:  fwflex.BoolValueToFramework(ctx, true),
			},
		}),
		AllowInterrupt: fwflex.BoolValueToFramework(ctx, true),
		AudioAndDTMFInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AudioAndDTMFInputSpecification{
			{
				StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 4000), //nolint:mnd
				AudioSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AudioSpecification{
					{
						EndTimeoutMs: fwflex.Int64ValueToFramework(ctx, 640),   //nolint:mnd
						MaxLengthMs:  fwflex.Int64ValueToFramework(ctx, 15000), //nolint:mnd
					},
				}),
				DTMFSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*DTMFSpecification{
					{
						DeletionCharacter: fwflex.StringValueToFramework(ctx, "*"),
						EndCharacter:      fwflex.StringValueToFramework(ctx, "#"),
						EndTimeoutMs:      fwflex.Int64ValueToFramework(ctx, 5000), //nolint:mnd
						MaxLength:         fwflex.Int64ValueToFramework(ctx, 513),  //nolint:mnd
					},
				}),
			},
		}),
		TextInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TextInputSpecification{
			{
				StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 30000), //nolint:mnd
			},
		}),
	}
}
