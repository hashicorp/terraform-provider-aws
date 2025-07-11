// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"
	"slices"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
)

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
