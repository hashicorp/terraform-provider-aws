// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwflex "github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

// confirmationSettingsEqualityFunc checks semantic equality.
// This specifically targets default values on PromptAttemptsSpecification
func confirmationSettingsEqualityFunc(ctx context.Context, oldValue, newValue fwtypes.NestedCollectionValue[IntentConfirmationSetting]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	// Convert old value to pointer
	oldConfirmationSettings, di := oldValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	// Convert new value to pointer
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

		return arePromptAttemptsEqual(ctx, oldPromptSpec.PromptAttemptsSpecification, newPromptSpec.PromptAttemptsSpecification)
	}

	return false, diags
}

func promptAttemptSpecificationSliceToMap(_ context.Context, input []*PromptAttemptsSpecification) map[string]*PromptAttemptsSpecification {
	output := make(map[string]*PromptAttemptsSpecification)
	for _, item := range input {
		if item == nil {
			continue
		}
		key := item.MapBlockKey.ValueString()
		output[key] = item
	}
	return output
}

// arePromptAttemptsEqual compares two PromptAttemptsSpecification fields for equality
// treating them as maps with map_block_key as the key
func arePromptAttemptsEqual(ctx context.Context, oldAttempts, newAttempts fwtypes.SetNestedObjectValueOf[PromptAttemptsSpecification]) (bool, diag.Diagnostics) {
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

		newPromptAttemptSpecMap := promptAttemptSpecificationSliceToMap(ctx, newPromptAttemptSpecification)

		var hasDefaults bool
		for _, value := range oldPromptAttemptSpecification {
			key := value.MapBlockKey.ValueString()
			if newValue, ok := newPromptAttemptSpecMap[key]; ok {
				// Compare the values for this key
				if !arePromptAttemptValuesEqual(*newValue, *value) {
					return false, diags
				}
			}

			if _, ok := promptAttemptSpecificationDefaults(ctx, key); ok {
				hasDefaults = true
				continue
			}
		}

		return hasDefaults, diags
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

// promptAttemptSpecificationDefaults returns the default PromptAttemptsSpecification for a given key
func promptAttemptSpecificationDefaults(ctx context.Context, key string) (*PromptAttemptsSpecification, bool) {
	objectDefault := &PromptAttemptsSpecification{
		MapBlockKey: fwtypes.StringEnumValue[PromptAttemptsType]("Initial"),
		AllowedInputTypes: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AllowedInputTypes{
			{
				AllowAudioInput: fwflex.BoolValueToFramework(ctx, true),
				AllowDTMFInput:  fwflex.BoolValueToFramework(ctx, true),
			},
		}),
		AllowInterrupt: fwflex.BoolValueToFramework(ctx, true),
		AudioAndDTMFInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AudioAndDTMFInputSpecification{
			{
				StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 4000),
				AudioSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*AudioSpecification{
					{
						EndTimeoutMs: fwflex.Int64ValueToFramework(ctx, 640),
						MaxLengthMs:  fwflex.Int64ValueToFramework(ctx, 15000),
					},
				}),
				DTMFSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*DTMFSpecification{
					{
						DeletionCharacter: fwflex.StringValueToFramework(ctx, "*"),
						EndCharacter:      fwflex.StringValueToFramework(ctx, "#"),
						EndTimeoutMs:      fwflex.Int64ValueToFramework(ctx, 5000),
						MaxLength:         fwflex.Int64ValueToFramework(ctx, 513),
					},
				}),
			},
		}),
		TextInputSpecification: fwtypes.NewListNestedObjectValueOfSliceMust(ctx, []*TextInputSpecification{
			{
				StartTimeoutMs: fwflex.Int64ValueToFramework(ctx, 30000),
			},
		}),
	}

	defaults := map[string]*PromptAttemptsSpecification{
		"Initial": objectDefault,
		"Retry":   objectDefault,
	}

	if val, ok := defaults[key]; ok {
		return val, true
	}

	return nil, false
}
