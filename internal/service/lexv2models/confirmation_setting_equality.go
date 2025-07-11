// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models

import (
	"context"

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
