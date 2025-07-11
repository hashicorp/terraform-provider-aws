// Copyright (c) HashiCorp, Inc.
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
					if !oldValueElicitationSetting.DefaultValueSpecification.Equal(newValueElicitationSetting.DefaultValueSpecification) ||
						!oldValueElicitationSetting.SampleUtterance.Equal(newValueElicitationSetting.SampleUtterance) ||
						!oldValueElicitationSetting.WaitAndContinueSpecification.Equal(newValueElicitationSetting.WaitAndContinueSpecification) {
						return false, diags
					}

					oldPromptSpec, di := oldValueElicitationSetting.PromptSpecification.ToPtr(ctx)
					diags = append(diags, di...)
					if diags.HasError() {
						return false, diags
					}

					newPromptSpec, di := newValueElicitationSetting.PromptSpecification.ToPtr(ctx)
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
			}
		}
	}

	return false, diags
}
