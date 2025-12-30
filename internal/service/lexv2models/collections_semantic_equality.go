// Copyright IBM Corp. 2014, 2025
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
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

		return evaluatePromptSpecification(ctx, oldConfirmationSettings.PromptSpecification, newConfirmationSettings.PromptSpecification)
	}

	return false, diags
}

// subSlotSettingEqualityFunc checks semantic equality.
// This specifically targets default values on PromptAttemptsSpecification
func subSlotSettingEqualityFunc(ctx context.Context, oldValue, newValue fwtypes.NestedCollectionValue[SubSlotSettingData]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldSubSlotSettings, di := oldValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	newSubSlotSettings, di := newValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	if oldSubSlotSettings != nil && newSubSlotSettings != nil {
		if !oldSubSlotSettings.Expression.Equal(newSubSlotSettings.Expression) {
			return false, diags
		}

		oldSlotSpecification, di := oldSubSlotSettings.SlotSpecification.ToSlice(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		newSlotSpecification, di := newSubSlotSettings.SlotSpecification.ToSlice(ctx)
		diags = append(diags, di...)
		if diags.HasError() {
			return false, diags
		}

		if oldSlotSpecification != nil && newSlotSpecification != nil {
			if slices.Equal(oldSlotSpecification, newSlotSpecification) {
				return true, diags
			}

			var promptsAreEquals bool
			for _, oldSlotSpec := range oldSlotSpecification {
				index := slices.IndexFunc(newSlotSpecification, func(item *SlotSpecificationsData) bool {
					return item.MapBlockKey.ValueString() == oldSlotSpec.MapBlockKey.ValueString()
				})

				if index == -1 {
					return false, diags
				}

				newSlotSpec := newSlotSpecification[index]
				if !oldSlotSpec.SlotTypeID.Equal(newSlotSpec.SlotTypeID) {
					return false, diags
				}

				oldValueElicitationSetting, di := oldSlotSpec.ValueElicitationSetting.ToPtr(ctx)
				diags = append(diags, di...)
				if diags.HasError() {
					return false, diags
				}

				newValueElicitationSetting, di := newSlotSpec.ValueElicitationSetting.ToPtr(ctx)
				diags = append(diags, di...)
				if diags.HasError() {
					return false, diags
				}

				if oldValueElicitationSetting != nil && newValueElicitationSetting != nil {
					out, di := evalValueElicitationSetting(ctx, oldValueElicitationSetting, newValueElicitationSetting)
					diags = append(diags, di...)
					if diags.HasError() {
						return false, diags
					}

					promptsAreEquals = out
				}
			}

			return promptsAreEquals, diags
		}
	}

	return false, diags
}

// valueElicitationSettingEqualityFunc checks semantic equality.
// This specifically targets default values on PromptAttemptsSpecification
func valueElicitationSettingEqualityFunc(ctx context.Context, oldValue, newValue fwtypes.NestedCollectionValue[ValueElicitationSettingData]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldValueElicitationSetting, di := oldValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	newValueElicitationSetting, di := newValue.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	if oldValueElicitationSetting != nil && newValueElicitationSetting != nil {
		return evalValueElicitationSetting(ctx, oldValueElicitationSetting, newValueElicitationSetting)
	}

	return false, diags
}

type valueElicitationSettinger interface {
	*ValueElicitationSettingData | *SubSlotValueElicitationSettingData
}

func evalValueElicitationSetting[T valueElicitationSettinger](ctx context.Context, oldValueElicitationSetting, newValueElicitationSetting T) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	switch oldSetting := any(oldValueElicitationSetting).(type) {
	case *ValueElicitationSettingData:
		newSetting := any(newValueElicitationSetting).(*ValueElicitationSettingData)
		if !oldSetting.DefaultValueSpecification.Equal(newSetting.DefaultValueSpecification) ||
			!oldSetting.SlotConstraint.Equal(newSetting.SlotConstraint) ||
			!oldSetting.SlotResolutionSetting.Equal(newSetting.SlotResolutionSetting) ||
			!oldSetting.SampleUtterance.Equal(newSetting.SampleUtterance) ||
			!oldSetting.WaitAndContinueSpecification.Equal(newSetting.WaitAndContinueSpecification) {
			return false, diags
		}

		return evaluatePromptSpecification(ctx, oldSetting.PromptSpecification, newSetting.PromptSpecification)
	case *SubSlotValueElicitationSettingData:
		newSetting := any(newValueElicitationSetting).(*SubSlotValueElicitationSettingData)
		if !oldSetting.DefaultValueSpecification.Equal(newSetting.DefaultValueSpecification) ||
			!oldSetting.SampleUtterance.Equal(newSetting.SampleUtterance) ||
			!oldSetting.WaitAndContinueSpecification.Equal(newSetting.WaitAndContinueSpecification) {
			return false, diags
		}

		return evaluatePromptSpecification(ctx, oldSetting.PromptSpecification, newSetting.PromptSpecification)
	}

	return false, diags
}

func evaluatePromptSpecification(ctx context.Context, oldSetting, newSetting fwtypes.ListNestedObjectValueOf[PromptSpecification]) (bool, diag.Diagnostics) {
	var diags diag.Diagnostics

	oldPromptSpec, di := oldSetting.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	newPromptSpec, di := newSetting.ToPtr(ctx)
	diags = append(diags, di...)
	if diags.HasError() {
		return false, diags
	}

	if oldPromptSpec != nil && newPromptSpec != nil {
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
