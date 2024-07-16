// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	lextypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/aws/smithy-go/middleware"
	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"github.com/hashicorp/terraform-plugin-framework-timetypes/timetypes"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/names"
)

// TestIntentAutoFlex is designed to extensively evaluate the capabilities of Intent expansion and
// flattening when utilizing the new feature called "autoflex." Given that autoflex is a recent
// addition and has been built upon the foundation of Intent, these unit tests play a crucial role
// in ensuring the reliability of the implementation.
//
// Looking ahead, for typical scenarios involving straightforward applications of autoflex's Expand
// and Flatten, it is generally unnecessary to conduct tests at the same level of detail as seen in
// this specific autoflex unit test. This guideline is applicable unless dealing with intricate
// resource schemas or situations where there is a genuine concern about the overall functionality.
// In such complex cases, it might still be advisable to perform thorough unit testing with
// autoflex to ensure everything functions as expected.
func TestIntentAutoFlex(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testString := "b72d06fd-2b78-5fe2-a6a3-e06e5efde347"
	testString2 := "a47c2004-f58b-5982-880a-f68c80f6307c"

	ssmlMessageTF := tflexv2models.SSMLMessage{
		Value: types.StringValue(testString),
	}
	ssmlMessageAWS := lextypes.SSMLMessage{
		Value: aws.String(testString),
	}

	plainTextMessageTF := tflexv2models.PlainTextMessage{
		Value: types.StringValue(testString),
	}
	plainTextMessageAWS := lextypes.PlainTextMessage{
		Value: aws.String(testString),
	}

	buttonTF := tflexv2models.Button{
		Text:  types.StringValue(testString),
		Value: types.StringValue(testString),
	}
	buttonAWS := lextypes.Button{
		Text:  aws.String(testString),
		Value: aws.String(testString),
	}

	buttonsTF := []tflexv2models.Button{
		buttonTF,
	}
	buttonsAWS := []lextypes.Button{
		buttonAWS,
	}

	imageResponseCardTF := tflexv2models.ImageResponseCard{
		Title:    types.StringValue(testString),
		Button:   fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.Button](ctx, buttonsTF),
		ImageURL: types.StringValue(testString),
		Subtitle: types.StringValue(testString),
	}
	imageResponseCardAWS := lextypes.ImageResponseCard{
		Title:    aws.String(testString),
		Buttons:  buttonsAWS,
		ImageUrl: aws.String(testString),
		Subtitle: aws.String(testString),
	}

	customPayloadTF := tflexv2models.CustomPayload{
		Value: types.StringValue(testString),
	}
	customPayloadAWS := lextypes.CustomPayload{
		Value: aws.String(testString),
	}

	messageTF := tflexv2models.Message{
		CustomPayload:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &customPayloadTF),
		ImageResponseCard: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &imageResponseCardTF),
		PlainTextMessage:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &plainTextMessageTF),
		SSMLMessage:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &ssmlMessageTF),
	}
	messageAWS := lextypes.Message{
		CustomPayload:     &customPayloadAWS,
		ImageResponseCard: &imageResponseCardAWS,
		PlainTextMessage:  &plainTextMessageAWS,
		SsmlMessage:       &ssmlMessageAWS,
	}

	messagesTF := []tflexv2models.Message{
		messageTF,
	}
	messagesAWS := []lextypes.Message{
		messageAWS,
	}

	messageGroupTF := tflexv2models.MessageGroup{
		Message:   fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &messageTF),
		Variation: fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.Message](ctx, messagesTF),
	}
	messageGroupAWS := []lextypes.MessageGroup{
		{
			Message:    &messageAWS,
			Variations: messagesAWS,
		},
	}

	responseSpecificationTF := tflexv2models.ResponseSpecification{
		MessageGroup:   fwtypes.NewListNestedObjectValueOfPtrMust[tflexv2models.MessageGroup](ctx, &messageGroupTF),
		AllowInterrupt: types.BoolValue(true),
	}
	responseSpecificationAWS := lextypes.ResponseSpecification{
		MessageGroups: []lextypes.MessageGroup{
			{
				Message: &messageAWS,
				Variations: []lextypes.Message{
					messageAWS,
				},
			},
		},
		AllowInterrupt: aws.Bool(true),
	}

	slotValueTF := tflexv2models.SlotValue{
		InterpretedValue: types.StringValue(testString),
	}
	slotValueAWS := lextypes.SlotValue{
		InterpretedValue: aws.String(testString),
	}

	slotValueOverrideAWS := lextypes.SlotValueOverride{
		Shape: lextypes.SlotShapeList,
		Value: &slotValueAWS,
		//Values: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []tflexv2models.SlotValueOverride{ // recursive so must be defined in line instead of in variable
		//	{
		//		Shape: types.StringValue(testString),
		//		Value: fwtypes.NewListNestedObjectValueOfPtr(ctx, &slotValueTF),
		//	},
		//}
	}

	slotValueOverrideMapTF := tflexv2models.SlotValueOverride{
		MapBlockKey: types.StringValue(testString),
		Shape:       fwtypes.StringEnumValue(lextypes.SlotShapeList),
		Value:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &slotValueTF),
		//Values: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []tflexv2models.SlotValueOverride{ // recursive so must be defined in line instead of in variable
		//	{
		//		Shape: types.StringValue(testString),
		//		Value: fwtypes.NewListNestedObjectValueOfPtr(ctx, &slotValueTF),
		//	},
		//}
	}
	slotValueOverrideMapAWS := map[string]lextypes.SlotValueOverride{
		testString: slotValueOverrideAWS,
	}

	intentOverrideTF := tflexv2models.IntentOverride{
		Name: types.StringValue(testString),
		Slot: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &slotValueOverrideMapTF),
	}
	intentOverrideAWS := lextypes.IntentOverride{
		Name:  aws.String(testString),
		Slots: slotValueOverrideMapAWS,
	}

	dialogActionTF := tflexv2models.DialogAction{
		Type:                fwtypes.StringEnumValue(lextypes.DialogActionTypeCloseIntent),
		SlotToElicit:        types.StringValue(testString),
		SuppressNextMessage: types.BoolValue(true),
	}
	dialogActionAWS := lextypes.DialogAction{
		Type:                lextypes.DialogActionTypeCloseIntent,
		SlotToElicit:        aws.String(testString),
		SuppressNextMessage: aws.Bool(true),
	}

	dialogStateTF := tflexv2models.DialogState{
		DialogAction: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogActionTF),
		Intent:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentOverrideTF),
		SessionAttributes: fwtypes.NewMapValueOfMust[basetypes.StringValue](ctx, map[string]attr.Value{
			testString: types.StringValue(testString2),
		}),
	}
	dialogStateAWS := lextypes.DialogState{
		DialogAction: &dialogActionAWS,
		Intent:       &intentOverrideAWS,
		SessionAttributes: map[string]string{
			testString: testString2,
		},
	}

	conditionTF := tflexv2models.Condition{
		ExpressionString: types.StringValue(testString),
	}

	defaultConditionalBranchTF := tflexv2models.DefaultConditionalBranch{
		NextStep: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		Response: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
	}

	conditionalSpecificationTF := tflexv2models.ConditionalSpecification{
		Active: types.BoolValue(true),
		ConditionalBranch: fwtypes.NewListNestedObjectValueOfValueSliceMust(ctx, []tflexv2models.ConditionalBranch{{
			Condition: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionTF),
			Name:      types.StringValue(testString),
			NextStep:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
			Response:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		}}),

		DefaultBranch: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &defaultConditionalBranchTF),
	}
	conditionalSpecificationAWS := lextypes.ConditionalSpecification{
		Active: aws.Bool(true),
		ConditionalBranches: []lextypes.ConditionalBranch{{
			Condition: &lextypes.Condition{
				ExpressionString: aws.String(testString),
			},
			Name:     aws.String(testString),
			NextStep: &dialogStateAWS,
			Response: &responseSpecificationAWS,
		}},
		DefaultBranch: &lextypes.DefaultConditionalBranch{
			NextStep: &dialogStateAWS,
			Response: &responseSpecificationAWS,
		},
	}

	intentClosingSettingTF := tflexv2models.IntentClosingSetting{
		Active:          types.BoolValue(true),
		ClosingResponse: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		Conditional:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		NextStep:        fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
	}
	intentClosingSettingAWS := lextypes.IntentClosingSetting{
		Active:          aws.Bool(true),
		ClosingResponse: &responseSpecificationAWS,
		Conditional:     &conditionalSpecificationAWS,
		NextStep:        &dialogStateAWS,
	}

	allowedInputTypesTF := tflexv2models.AllowedInputTypes{
		AllowAudioInput: types.BoolValue(true),
		AllowDTMFInput:  types.BoolValue(true),
	}
	allowedInputTypesAWS := lextypes.AllowedInputTypes{
		AllowAudioInput: aws.Bool(true),
		AllowDTMFInput:  aws.Bool(true),
	}

	audioSpecificationTF := tflexv2models.AudioSpecification{
		EndTimeoutMs: types.Int64Value(1),
		MaxLengthMs:  types.Int64Value(1),
	}
	audioSpecificationAWS := lextypes.AudioSpecification{
		EndTimeoutMs: aws.Int32(1),
		MaxLengthMs:  aws.Int32(1),
	}

	dtmfSpecificationTF := tflexv2models.DTMFSpecification{
		DeletionCharacter: types.StringValue(testString),
		EndCharacter:      types.StringValue(testString),
		EndTimeoutMs:      types.Int64Value(1),
		MaxLength:         types.Int64Value(1),
	}
	dtmfSpecificationAWS := lextypes.DTMFSpecification{
		DeletionCharacter: aws.String(testString),
		EndCharacter:      aws.String(testString),
		EndTimeoutMs:      aws.Int32(1),
		MaxLength:         aws.Int32(1),
	}

	audioAndDTMFInputSpecificationTF := tflexv2models.AudioAndDTMFInputSpecification{
		StartTimeoutMs:     types.Int64Value(1),
		AudioSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &audioSpecificationTF),
		DTMFSpecification:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dtmfSpecificationTF),
	}
	audioAndDTMFInputSpecificationAWS := lextypes.AudioAndDTMFInputSpecification{
		StartTimeoutMs:     aws.Int32(1),
		AudioSpecification: &audioSpecificationAWS,
		DtmfSpecification:  &dtmfSpecificationAWS,
	}

	textInputSpecificationTF := tflexv2models.TextInputSpecification{
		StartTimeoutMs: types.Int64Value(1),
	}
	textInputSpecificationAWS := lextypes.TextInputSpecification{
		StartTimeoutMs: aws.Int32(1),
	}

	promptAttemptSpecificationTF := tflexv2models.PromptAttemptsSpecification{
		MapBlockKey:                    fwtypes.StringEnumValue(tflexv2models.PromptAttemptsTypeInitial),
		AllowedInputTypes:              fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &allowedInputTypesTF),
		AllowInterrupt:                 types.BoolValue(true),
		AudioAndDTMFInputSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &audioAndDTMFInputSpecificationTF),
		TextInputSpecification:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &textInputSpecificationTF),
	}
	promptAttemptSpecificationAWS := lextypes.PromptAttemptSpecification{
		AllowedInputTypes:              &allowedInputTypesAWS,
		AllowInterrupt:                 aws.Bool(true),
		AudioAndDTMFInputSpecification: &audioAndDTMFInputSpecificationAWS,
		TextInputSpecification:         &textInputSpecificationAWS,
	}

	promptSpecificationTF := tflexv2models.PromptSpecification{
		MaxRetries:                  types.Int64Value(1),
		MessageGroup:                fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &messageGroupTF),
		AllowInterrupt:              types.BoolValue(true),
		MessageSelectionStrategy:    fwtypes.StringEnumValue(lextypes.MessageSelectionStrategyOrdered),
		PromptAttemptsSpecification: fwtypes.NewSetNestedObjectValueOfPtrMust(ctx, &promptAttemptSpecificationTF),
	}
	promptSpecificationAWS := lextypes.PromptSpecification{
		MaxRetries:               aws.Int32(1),
		MessageGroups:            messageGroupAWS,
		AllowInterrupt:           aws.Bool(true),
		MessageSelectionStrategy: lextypes.MessageSelectionStrategyOrdered,
		PromptAttemptsSpecification: map[string]lextypes.PromptAttemptSpecification{
			string(tflexv2models.PromptAttemptsTypeInitial): promptAttemptSpecificationAWS,
		},
	}

	failureSuccessTimeoutTF := tflexv2models.FailureSuccessTimeout{
		FailureConditional: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		FailureNextStep:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		FailureResponse:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		SuccessConditional: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		SuccessNextStep:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		SuccessResponse:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		TimeoutConditional: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		TimeoutNextStep:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		TimeoutResponse:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
	}

	postCodeHookSpecificationAWS := lextypes.PostDialogCodeHookInvocationSpecification{
		FailureConditional: &conditionalSpecificationAWS,
		FailureNextStep:    &dialogStateAWS,
		FailureResponse:    &responseSpecificationAWS,
		SuccessConditional: &conditionalSpecificationAWS,
		SuccessNextStep:    &dialogStateAWS,
		SuccessResponse:    &responseSpecificationAWS,
		TimeoutConditional: &conditionalSpecificationAWS,
		TimeoutNextStep:    &dialogStateAWS,
		TimeoutResponse:    &responseSpecificationAWS,
	}
	postFulfillmentStatusSpecificationAWS := lextypes.PostFulfillmentStatusSpecification{
		FailureConditional: &conditionalSpecificationAWS,
		FailureNextStep:    &dialogStateAWS,
		FailureResponse:    &responseSpecificationAWS,
		SuccessConditional: &conditionalSpecificationAWS,
		SuccessNextStep:    &dialogStateAWS,
		SuccessResponse:    &responseSpecificationAWS,
		TimeoutConditional: &conditionalSpecificationAWS,
		TimeoutNextStep:    &dialogStateAWS,
		TimeoutResponse:    &responseSpecificationAWS,
	}

	dialogCodeHookInvocationSettingTF := tflexv2models.DialogCodeHookInvocationSetting{
		Active:                    types.BoolValue(true),
		EnableCodeHookInvocation:  types.BoolValue(true),
		InvocationLabel:           types.StringValue(testString),
		PostCodeHookSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &failureSuccessTimeoutTF),
	}
	dialogCodeHookInvocationSettingAWS := lextypes.DialogCodeHookInvocationSetting{
		Active:                    aws.Bool(true),
		EnableCodeHookInvocation:  aws.Bool(true),
		InvocationLabel:           aws.String(testString),
		PostCodeHookSpecification: &postCodeHookSpecificationAWS,
	}

	elicitationCodeHookTF := tflexv2models.ElicitationCodeHookInvocationSetting{
		EnableCodeHookInvocation: types.BoolValue(true),
		InvocationLabel:          types.StringValue(testString),
	}
	elicitationCodeHookAWS := lextypes.ElicitationCodeHookInvocationSetting{
		EnableCodeHookInvocation: aws.Bool(true),
		InvocationLabel:          aws.String(testString),
	}

	intentConfirmationSettingTF := tflexv2models.IntentConfirmationSetting{
		PromptSpecification:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &promptSpecificationTF),
		Active:                  types.BoolValue(true),
		CodeHook:                fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogCodeHookInvocationSettingTF),
		ConfirmationConditional: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		ConfirmationNextStep:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		ConfirmationResponse:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		DeclinationConditional:  fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		DeclinationNextStep:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		DeclinationResponse:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		ElicitationCodeHook:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &elicitationCodeHookTF),
		FailureConditional:      fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		FailureNextStep:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
		FailureResponse:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
	}
	intentConfirmationSettingAWS := lextypes.IntentConfirmationSetting{
		PromptSpecification:     &promptSpecificationAWS,
		Active:                  aws.Bool(true),
		CodeHook:                &dialogCodeHookInvocationSettingAWS,
		ConfirmationConditional: &conditionalSpecificationAWS,
		ConfirmationNextStep:    &dialogStateAWS,
		ConfirmationResponse:    &responseSpecificationAWS,
		DeclinationConditional:  &conditionalSpecificationAWS,
		DeclinationNextStep:     &dialogStateAWS,
		DeclinationResponse:     &responseSpecificationAWS,
		ElicitationCodeHook:     &elicitationCodeHookAWS,
		FailureConditional:      &conditionalSpecificationAWS,
		FailureNextStep:         &dialogStateAWS,
		FailureResponse:         &responseSpecificationAWS,
	}

	dialogCodeHookSettingsTF := tflexv2models.DialogCodeHookSettings{
		Enabled: types.BoolValue(true),
	}
	dialogCodeHookSettingsAWS := lextypes.DialogCodeHookSettings{
		Enabled: true,
	}

	fulfillmentStartResponseSpecificationTF := tflexv2models.FulfillmentStartResponseSpecification{
		DelayInSeconds: types.Int64Value(1),
		MessageGroup:   fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &messageGroupTF),
		AllowInterrupt: types.BoolValue(true),
	}
	fulfillmentStartResponseSpecificationAWS := lextypes.FulfillmentStartResponseSpecification{
		DelayInSeconds: aws.Int32(1),
		MessageGroups:  messageGroupAWS,
		AllowInterrupt: aws.Bool(true),
	}

	fulfillmentUpdateResponseSpecificationTF := tflexv2models.FulfillmentUpdateResponseSpecification{
		FrequencyInSeconds: types.Int64Value(1),
		MessageGroup:       fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &messageGroupTF),
		AllowInterrupt:     types.BoolValue(true),
	}
	fulfillmentUpdateResponseSpecificationAWS := lextypes.FulfillmentUpdateResponseSpecification{
		FrequencyInSeconds: aws.Int32(1),
		MessageGroups:      messageGroupAWS,
		AllowInterrupt:     aws.Bool(true),
	}

	fulfillmentUpdatesSpecificationTF := tflexv2models.FulfillmentUpdatesSpecification{
		Active:           types.BoolValue(true),
		StartResponse:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentStartResponseSpecificationTF),
		TimeoutInSeconds: types.Int64Value(1),
		UpdateResponse:   fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentUpdateResponseSpecificationTF),
	}
	fulfillmentUpdatesSpecificationAWS := lextypes.FulfillmentUpdatesSpecification{
		Active:           aws.Bool(true),
		StartResponse:    &fulfillmentStartResponseSpecificationAWS,
		TimeoutInSeconds: aws.Int32(1),
		UpdateResponse:   &fulfillmentUpdateResponseSpecificationAWS,
	}

	fulfillmentCodeHookSettingsTF := tflexv2models.FulfillmentCodeHookSettings{
		Enabled:                            types.BoolValue(true),
		Active:                             types.BoolValue(true),
		FulfillmentUpdatesSpecification:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentUpdatesSpecificationTF),
		PostFulfillmentStatusSpecification: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &failureSuccessTimeoutTF),
	}
	fulfillmentCodeHookSettingsAWS := lextypes.FulfillmentCodeHookSettings{
		Enabled:                            true,
		Active:                             aws.Bool(true),
		FulfillmentUpdatesSpecification:    &fulfillmentUpdatesSpecificationAWS,
		PostFulfillmentStatusSpecification: &postFulfillmentStatusSpecificationAWS,
	}

	initialResponseSettingTF := tflexv2models.InitialResponseSetting{
		CodeHook:        fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogCodeHookInvocationSettingTF),
		Conditional:     fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &conditionalSpecificationTF),
		InitialResponse: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &responseSpecificationTF),
		NextStep:        fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogStateTF),
	}
	initialResponseSettingAWS := lextypes.InitialResponseSetting{
		CodeHook:        &dialogCodeHookInvocationSettingAWS,
		Conditional:     &conditionalSpecificationAWS,
		InitialResponse: &responseSpecificationAWS,
		NextStep:        &dialogStateAWS,
	}

	inputContextTF := tflexv2models.InputContext{
		Name: types.StringValue(testString),
	}
	inputContextAWS := lextypes.InputContext{
		Name: aws.String(testString),
	}

	inputContextsTF := []tflexv2models.InputContext{
		inputContextTF,
	}
	inputContextsAWS := []lextypes.InputContext{
		inputContextAWS,
	}

	kendraConfigurationTF := tflexv2models.KendraConfiguration{
		KendraIndex:              types.StringValue(testString),
		QueryFilterString:        types.StringValue(testString),
		QueryFilterStringEnabled: types.BoolValue(true),
	}
	kendraConfigurationAWS := lextypes.KendraConfiguration{
		KendraIndex:              aws.String(testString),
		QueryFilterString:        aws.String(testString),
		QueryFilterStringEnabled: true,
	}

	outputContextTF := tflexv2models.OutputContext{
		Name:                types.StringValue(testString),
		TimeToLiveInSeconds: types.Int64Value(1),
		TurnsToLive:         types.Int64Value(1),
	}
	outputContextAWS := lextypes.OutputContext{
		Name:                aws.String(testString),
		TimeToLiveInSeconds: aws.Int32(1),
		TurnsToLive:         aws.Int32(1),
	}

	outputContextsTF := []tflexv2models.OutputContext{
		outputContextTF,
	}
	outputContextsAWS := []lextypes.OutputContext{
		outputContextAWS,
	}

	sampleUtteranceTF := tflexv2models.SampleUtterance{
		Utterance: types.StringValue(testString),
	}
	sampleUtteranceAWS := lextypes.SampleUtterance{
		Utterance: aws.String(testString),
	}

	sampleUtterancesTF := []tflexv2models.SampleUtterance{
		sampleUtteranceTF,
	}
	sampleUtterancesAWS := []lextypes.SampleUtterance{
		sampleUtteranceAWS,
	}

	slotPriorityTF := tflexv2models.SlotPriority{
		Priority: types.Int64Value(1),
		SlotID:   types.StringValue(testString),
	}
	slotPriorityAWS := lextypes.SlotPriority{
		Priority: aws.Int32(1),
		SlotId:   aws.String(testString),
	}

	slotPrioritiesTF := []tflexv2models.SlotPriority{
		slotPriorityTF,
	}
	slotPrioritiesAWS := []lextypes.SlotPriority{
		slotPriorityAWS,
	}

	intentCreateTF := tflexv2models.ResourceIntentData{
		BotID:                  types.StringValue(testString),
		BotVersion:             types.StringValue(testString),
		Name:                   types.StringValue(testString),
		LocaleID:               types.StringValue(testString),
		Description:            types.StringValue(testString),
		DialogCodeHook:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogCodeHookSettingsTF),
		FulfillmentCodeHook:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentCodeHookSettingsTF),
		InitialResponseSetting: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &initialResponseSettingTF),
		InputContext:           fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.InputContext](ctx, inputContextsTF),
		ClosingSetting:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentClosingSettingTF),
		ConfirmationSetting:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentConfirmationSettingTF),
		KendraConfiguration:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &kendraConfigurationTF),
		OutputContext:          fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.OutputContext](ctx, outputContextsTF),
		ParentIntentSignature:  types.StringValue(testString),
		SampleUtterance:        fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.SampleUtterance](ctx, sampleUtterancesTF),
	}
	intentCreateAWS := lexmodelsv2.CreateIntentInput{
		BotId:                     aws.String(testString),
		BotVersion:                aws.String(testString),
		IntentName:                aws.String(testString),
		LocaleId:                  aws.String(testString),
		Description:               aws.String(testString),
		DialogCodeHook:            &dialogCodeHookSettingsAWS,
		FulfillmentCodeHook:       &fulfillmentCodeHookSettingsAWS,
		InitialResponseSetting:    &initialResponseSettingAWS,
		InputContexts:             inputContextsAWS,
		IntentClosingSetting:      &intentClosingSettingAWS,
		IntentConfirmationSetting: &intentConfirmationSettingAWS,
		KendraConfiguration:       &kendraConfigurationAWS,
		OutputContexts:            outputContextsAWS,
		ParentIntentSignature:     aws.String(testString),
		SampleUtterances:          sampleUtterancesAWS,
	}

	intentModifyTF := tflexv2models.ResourceIntentData{
		BotID:                  types.StringValue(testString),
		BotVersion:             types.StringValue(testString),
		Name:                   types.StringValue(testString),
		LocaleID:               types.StringValue(testString),
		Description:            types.StringValue(testString),
		DialogCodeHook:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogCodeHookSettingsTF),
		FulfillmentCodeHook:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentCodeHookSettingsTF),
		InitialResponseSetting: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &initialResponseSettingTF),
		InputContext:           fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.InputContext](ctx, inputContextsTF),
		ClosingSetting:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentClosingSettingTF),
		ConfirmationSetting:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentConfirmationSettingTF),
		KendraConfiguration:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &kendraConfigurationTF),
		OutputContext:          fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.OutputContext](ctx, outputContextsTF),
		ParentIntentSignature:  types.StringValue(testString),
		SampleUtterance:        fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.SampleUtterance](ctx, sampleUtterancesTF),
		SlotPriority:           fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.SlotPriority](ctx, slotPrioritiesTF),
	}
	intentModifyAWS := lexmodelsv2.UpdateIntentInput{
		BotId:                     aws.String(testString),
		BotVersion:                aws.String(testString),
		IntentName:                aws.String(testString),
		LocaleId:                  aws.String(testString),
		Description:               aws.String(testString),
		DialogCodeHook:            &dialogCodeHookSettingsAWS,
		FulfillmentCodeHook:       &fulfillmentCodeHookSettingsAWS,
		InitialResponseSetting:    &initialResponseSettingAWS,
		InputContexts:             inputContextsAWS,
		IntentClosingSetting:      &intentClosingSettingAWS,
		IntentConfirmationSetting: &intentConfirmationSettingAWS,
		KendraConfiguration:       &kendraConfigurationAWS,
		OutputContexts:            outputContextsAWS,
		ParentIntentSignature:     aws.String(testString),
		SampleUtterances:          sampleUtterancesAWS,
		SlotPriorities:            slotPrioritiesAWS,
	}

	testTimeStr := "2023-12-08T09:34:01Z"
	testTimeTime := errs.Must(time.Parse(time.RFC3339, testTimeStr))

	intentDescribeTF := tflexv2models.ResourceIntentData{
		BotID:                  types.StringValue(testString),
		BotVersion:             types.StringValue(testString),
		ClosingSetting:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentClosingSettingTF),
		ConfirmationSetting:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &intentConfirmationSettingTF),
		CreationDateTime:       timetypes.NewRFC3339ValueMust(testTimeStr),
		Description:            types.StringValue(testString),
		DialogCodeHook:         fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &dialogCodeHookSettingsTF),
		FulfillmentCodeHook:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &fulfillmentCodeHookSettingsTF),
		IntentID:               types.StringValue(testString),
		InitialResponseSetting: fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &initialResponseSettingTF),
		InputContext:           fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.InputContext](ctx, inputContextsTF),
		KendraConfiguration:    fwtypes.NewListNestedObjectValueOfPtrMust(ctx, &kendraConfigurationTF),
		LastUpdatedDateTime:    timetypes.NewRFC3339ValueMust(testTimeStr),
		LocaleID:               types.StringValue(testString),
		Name:                   types.StringValue(testString),
		OutputContext:          fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.OutputContext](ctx, outputContextsTF),
		ParentIntentSignature:  types.StringValue(testString),
		SampleUtterance:        fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.SampleUtterance](ctx, sampleUtterancesTF),
		SlotPriority:           fwtypes.NewListNestedObjectValueOfValueSliceMust[tflexv2models.SlotPriority](ctx, slotPrioritiesTF),
	}
	intentDescribeAWS := lexmodelsv2.DescribeIntentOutput{
		BotId:                     aws.String(testString),
		BotVersion:                aws.String(testString),
		CreationDateTime:          aws.Time(testTimeTime),
		Description:               aws.String(testString),
		DialogCodeHook:            &dialogCodeHookSettingsAWS,
		FulfillmentCodeHook:       &fulfillmentCodeHookSettingsAWS,
		InitialResponseSetting:    &initialResponseSettingAWS,
		InputContexts:             inputContextsAWS,
		IntentClosingSetting:      &intentClosingSettingAWS,
		IntentConfirmationSetting: &intentConfirmationSettingAWS,
		IntentId:                  aws.String(testString),
		IntentName:                aws.String(testString),
		KendraConfiguration:       &kendraConfigurationAWS,
		LastUpdatedDateTime:       aws.Time(testTimeTime),
		LocaleId:                  aws.String(testString),
		OutputContexts:            outputContextsAWS,
		ParentIntentSignature:     aws.String(testString),
		SampleUtterances:          sampleUtterancesAWS,
		SlotPriorities:            slotPrioritiesAWS,
	}

	testCases := []struct {
		TestName string
		TFFull   any
		AWSFull  any
		TFEmpty  any
		AWSEmpty any
		WantErr  bool
	}{
		{
			TestName: names.AttrMessage,
			TFFull:   &messageTF,
			TFEmpty:  &tflexv2models.Message{},
			AWSFull:  &messageAWS,
			AWSEmpty: &lextypes.Message{},
		},
		{
			TestName: "responseSpecification",
			TFFull:   &responseSpecificationTF,
			TFEmpty:  &tflexv2models.ResponseSpecification{},
			AWSFull:  &responseSpecificationAWS,
			AWSEmpty: &lextypes.ResponseSpecification{},
		},
		{
			TestName: "promptSpecification",
			TFFull:   &promptSpecificationTF,
			TFEmpty:  &tflexv2models.PromptSpecification{},
			AWSFull:  &promptSpecificationAWS,
			AWSEmpty: &lextypes.PromptSpecification{},
		},
		{
			TestName: "dialogState",
			TFFull:   &dialogStateTF,
			TFEmpty:  &tflexv2models.DialogState{},
			AWSFull:  &dialogStateAWS,
			AWSEmpty: &lextypes.DialogState{},
		},
		{
			TestName: "dialogAction",
			TFFull:   &dialogActionTF,
			TFEmpty:  &tflexv2models.DialogAction{},
			AWSFull:  &dialogActionAWS,
			AWSEmpty: &lextypes.DialogAction{},
		},
		{
			TestName: "conditionalSpecification",
			TFFull:   &conditionalSpecificationTF,
			TFEmpty:  &tflexv2models.ConditionalSpecification{},
			AWSFull:  &conditionalSpecificationAWS,
			AWSEmpty: &lextypes.ConditionalSpecification{},
		},
		{
			TestName: "intentClosingSetting",
			TFFull:   &intentClosingSettingTF,
			TFEmpty:  &tflexv2models.IntentClosingSetting{},
			AWSFull:  &intentClosingSettingAWS,
			AWSEmpty: &lextypes.IntentClosingSetting{},
		},
		{
			TestName: "intentConfirmationSetting",
			TFFull:   &intentConfirmationSettingTF,
			TFEmpty:  &tflexv2models.IntentConfirmationSetting{},
			AWSFull:  &intentConfirmationSettingAWS,
			AWSEmpty: &lextypes.IntentConfirmationSetting{},
		},
		{
			TestName: "intentOverride",
			TFFull:   &intentOverrideTF,
			TFEmpty:  &tflexv2models.IntentOverride{},
			AWSFull:  &intentOverrideAWS,
			AWSEmpty: &lextypes.IntentOverride{},
		},
		{
			TestName: "slotValue",
			TFFull:   &slotValueTF,
			TFEmpty:  &tflexv2models.SlotValue{},
			AWSFull:  &slotValueAWS,
			AWSEmpty: &lextypes.SlotValue{},
		},
		{
			TestName: "create intent",
			TFFull:   &intentCreateTF,
			TFEmpty:  &tflexv2models.ResourceIntentData{},
			AWSFull:  &intentCreateAWS,
			AWSEmpty: &lexmodelsv2.CreateIntentInput{},
		},
		{
			TestName: "update intent",
			TFFull:   &intentModifyTF,
			TFEmpty:  &tflexv2models.ResourceIntentData{},
			AWSFull:  &intentModifyAWS,
			AWSEmpty: &lexmodelsv2.UpdateIntentInput{},
		},
		{
			TestName: "describe intent",
			TFFull:   &intentDescribeTF,
			TFEmpty:  &tflexv2models.ResourceIntentData{},
			AWSFull:  &intentDescribeAWS,
			AWSEmpty: &lexmodelsv2.DescribeIntentOutput{},
		},
	}

	ignoreExpoOpts := cmpopts.IgnoreUnexported(
		lexmodelsv2.CreateIntentInput{},
		lexmodelsv2.DescribeIntentOutput{},
		lexmodelsv2.UpdateIntentInput{},
		lextypes.AllowedInputTypes{},
		lextypes.AudioAndDTMFInputSpecification{},
		lextypes.AudioSpecification{},
		lextypes.Button{},
		lextypes.Condition{},
		lextypes.ConditionalBranch{},
		lextypes.ConditionalSpecification{},
		lextypes.CustomPayload{},
		lextypes.DefaultConditionalBranch{},
		lextypes.DialogAction{},
		lextypes.DialogCodeHookInvocationSetting{},
		lextypes.DialogCodeHookSettings{},
		lextypes.DialogState{},
		lextypes.DTMFSpecification{},
		lextypes.ElicitationCodeHookInvocationSetting{},
		lextypes.FulfillmentCodeHookSettings{},
		lextypes.FulfillmentStartResponseSpecification{},
		lextypes.FulfillmentUpdateResponseSpecification{},
		lextypes.FulfillmentUpdatesSpecification{},
		lextypes.ImageResponseCard{},
		lextypes.InitialResponseSetting{},
		lextypes.InputContext{},
		lextypes.IntentClosingSetting{},
		lextypes.IntentConfirmationSetting{},
		lextypes.IntentOverride{},
		lextypes.KendraConfiguration{},
		lextypes.Message{},
		lextypes.MessageGroup{},
		lextypes.OutputContext{},
		lextypes.PlainTextMessage{},
		lextypes.PostDialogCodeHookInvocationSpecification{},
		lextypes.PostFulfillmentStatusSpecification{},
		lextypes.PromptAttemptSpecification{},
		lextypes.PromptSpecification{},
		lextypes.ResponseSpecification{},
		lextypes.SampleUtterance{},
		lextypes.SlotPriority{},
		lextypes.SlotValue{},
		lextypes.SlotValueOverride{},
		lextypes.SSMLMessage{},
		lextypes.TextInputSpecification{},
		middleware.Metadata{},
	)

	for _, testCase := range testCases {
		testCase := testCase

		t.Run(fmt.Sprintf("expand %s", testCase.TestName), func(t *testing.T) {
			t.Parallel()

			diags := flex.Expand(context.WithValue(ctx, flex.ResourcePrefix, "Intent"), testCase.TFFull, testCase.AWSEmpty)

			gotErr := diags != nil

			if gotErr != testCase.WantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.WantErr)
			}

			if gotErr {
				if !testCase.WantErr {
					t.Errorf("err = %q", diags)
				}
			} else {
				if diff := cmp.Diff(testCase.AWSEmpty, testCase.AWSFull, ignoreExpoOpts); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})

		t.Run(fmt.Sprintf("flatten %s", testCase.TestName), func(t *testing.T) {
			t.Parallel()

			diags := flex.Flatten(context.WithValue(ctx, flex.ResourcePrefix, "Intent"), testCase.AWSFull, testCase.TFEmpty)

			gotErr := diags != nil

			if gotErr != testCase.WantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.WantErr)
			}

			if gotErr {
				if !testCase.WantErr {
					t.Errorf("err = %q", diags)
				}
			} else {
				// because TF type has .Equal method, cmp can act strangely - string comparison shortcut
				// avoids
				if fmt.Sprint(testCase.TFEmpty) == fmt.Sprint(testCase.TFFull) {
					return
				}

				if diff := cmp.Diff(testCase.TFEmpty, testCase.TFFull, ignoreExpoOpts); diff != "" {
					t.Errorf("unexpected diff (+wanted, -got): %s", diff)
				}
			}
		})
	}
}

// Acceptance tests access AWS and cost money to run.

func TestAccLexV2ModelsIntent_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexV2ModelsIntent_disappears(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceIntent, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateConfirmationSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateConfirmationSetting(rName, 1, "test", 640, 640),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*", map[string]string{
						"max_retries":                acctest.Ct1,
						"message_selection_strategy": "Ordered",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*.message_group.*.message.*.plain_text_message.*", map[string]string{
						names.AttrValue: "test",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*.prompt_attempts_specification.*.audio_and_dtmf_input_specification.*.audio_specification.*", map[string]string{
						"end_timeout_ms": "640",
					}),
				),
			},
			{
				Config: testAccIntentConfig_updateConfirmationSetting(rName, 1, "test", 650, 660),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*", map[string]string{
						"max_retries":                acctest.Ct1,
						"message_selection_strategy": "Ordered",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*.message_group.*.message.*.plain_text_message.*", map[string]string{
						names.AttrValue: "test",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*.prompt_attempts_specification.*.audio_and_dtmf_input_specification.*.audio_specification.*", map[string]string{
						"end_timeout_ms": "650",
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "confirmation_setting.*.prompt_specification.*.prompt_attempts_specification.*.audio_and_dtmf_input_specification.*.audio_specification.*", map[string]string{
						"end_timeout_ms": "660",
					}),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateClosingSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateClosingSetting(rName, "test1", "test2", "test3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "closing_setting.*", map[string]string{
						"active": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "closing_setting.*.conditional.*.conditional_branch.*", map[string]string{
						names.AttrName: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.closing_response.0.message_group.0.message.0.plain_text_message.0.value", "test1"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "test2"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "test3"),
				),
			},
			{
				Config: testAccIntentConfig_updateClosingSetting(rName, "Hvad", "er", "hygge"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "closing_setting.*", map[string]string{
						"active": acctest.CtTrue,
					}),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "closing_setting.*.conditional.*.conditional_branch.*", map[string]string{
						names.AttrName: rName,
					}),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.closing_response.0.message_group.0.message.0.plain_text_message.0.value", "Hvad"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "er"),
					resource.TestCheckResourceAttr(resourceName, "closing_setting.0.conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "hygge"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateInputContext(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateInputContext(rName, "sammanhang1", "sammanhang2", "sammanhang3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "input_context.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "input_context.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.0.name", "sammanhang1"),
					resource.TestCheckResourceAttr(resourceName, "input_context.1.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.1.name", "sammanhang2"),
					resource.TestCheckResourceAttr(resourceName, "input_context.2.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.2.name", "sammanhang3"),
				),
			},
			{
				Config: testAccIntentConfig_updateInputContext(rName, "kropp", "utan", "blod"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "input_context.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "input_context.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.0.name", "kropp"),
					resource.TestCheckResourceAttr(resourceName, "input_context.1.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.1.name", "utan"),
					resource.TestCheckResourceAttr(resourceName, "input_context.2.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "input_context.2.name", "blod"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateInitialResponseSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateInitialResponseSetting(rName, "branch1", "tre", "slumpmässiga", "ord"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.enable_code_hook_invocation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.invocation_label", "test"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.0.expression_string", "slot1 = \"test\""),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.name", "branch1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "tre"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "slumpmässiga"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.0.value", "ord"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.next_step.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIntentConfig_updateInitialResponseSetting(rName, "gren1", "några", "olika", "bokstäver"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.enable_code_hook_invocation", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.invocation_label", "test"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.condition.0.expression_string", "slot1 = \"test\""),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.name", "gren1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "några"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "olika"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.failure_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.success_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.code_hook.0.post_code_hook_specification.0.timeout_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.plain_text_message.0.value", "bokstäver"),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.initial_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "initial_response_setting.0.next_step.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateFulfillmentCodeHook(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateFulfillmentCodeHook(rName, "meddelande", 10, "slumpmässiga", "gren1", "alfanumerisk", "olika"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.delay_in_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.0.value", "meddelande"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.timeout_in_seconds", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.frequency_in_seconds", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.0.value", "slumpmässiga"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.0.expression_string", "slot1 = \"test\""),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.name", "gren1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.suppress_next_message", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.map_block_key", "test"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.shape", "Scalar"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.0.interpreted_value", "alfanumerisk"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot2", "roligt2"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "olika"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "safriduo"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_response.#", acctest.Ct0),
				),
			},
			{
				Config: testAccIntentConfig_updateFulfillmentCodeHook(rName, "dagdröm", 10, "dansa", "dumbom", "gås", "mat"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.enabled", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.delay_in_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.plain_text_message.0.value", "dagdröm"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.start_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.timeout_in_seconds", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.frequency_in_seconds", acctest.Ct10),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.plain_text_message.0.value", "dansa"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.fulfillment_updates_specification.0.update_response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.%", "9"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.active", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.condition.0.expression_string", "slot1 = \"test\""),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.name", "dumbom"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.suppress_next_message", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.name", "test"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.map_block_key", "test"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.shape", "Scalar"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.intent.0.slot.0.value.0.interpreted_value", "gås"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot1", "roligt"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.next_step.0.session_attributes.slot2", "roligt2"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "mat"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.conditional_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.slot_to_elicit", "slot1"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.dialog_action.0.type", "CloseIntent"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.intent.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.next_step.0.session_attributes.slot1", "hallon"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.allow_interrupt", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.%", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.custom_payload.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.image_response_card.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.plain_text_message.0.value", "safriduo"),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.message.0.ssml_message.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_conditional.0.default_branch.0.response.0.message_group.0.variation.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.failure_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.success_response.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_conditional.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_next_step.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "fulfillment_code_hook.0.post_fulfillment_status_specification.0.timeout_response.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateDialogCodeHook(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateDialogCodeHook(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.0.enabled", acctest.CtTrue),
				),
			},
			{
				Config: testAccIntentConfig_updateDialogCodeHook(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "dialog_code_hook.0.enabled", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateOutputContext(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateOutputContext(rName, "name1", "name2", "name3"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "output_context.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.name", "name1"),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.turns_to_live", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.turns_to_live", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.name", "name3"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.turns_to_live", "5"),
				),
			},
			{
				Config: testAccIntentConfig_updateOutputContext(rName, "name2", "name3", "name4"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "output_context.#", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.name", "name2"),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.0.turns_to_live", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.name", "name3"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.1.turns_to_live", "5"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.name", "name4"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.time_to_live_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "output_context.2.turns_to_live", "5"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsIntent_updateSampleUtterance(t *testing.T) {
	ctx := acctest.Context(t)

	var intent lexmodelsv2.DescribeIntentOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_updateSampleUtterance(rName, "yttrande", "twocolors", "danny", "dansa"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.0.utterance", "yttrande"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.1.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.1.utterance", "twocolors"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.2.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.2.utterance", "danny"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.3.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.3.utterance", "dansa"),
				),
			},
			{
				Config: testAccIntentConfig_updateSampleUtterance(rName, "rustedroot", "sendme", "onmy", "way"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.#", acctest.Ct4),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.0.utterance", "rustedroot"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.1.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.1.utterance", "sendme"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.2.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.2.utterance", "onmy"),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.3.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "sample_utterance.3.utterance", "way"),
				),
			},
		},
	})
}

func testAccCheckIntentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_intent" {
				continue
			}

			_, err := conn.DescribeIntent(ctx, &lexmodelsv2.DescribeIntentInput{
				IntentId:   aws.String(rs.Primary.Attributes["intent_id"]),
				BotId:      aws.String(rs.Primary.Attributes["bot_id"]),
				BotVersion: aws.String(rs.Primary.Attributes["bot_version"]),
				LocaleId:   aws.String(rs.Primary.Attributes["locale_id"]),
			})
			if errs.IsA[*lextypes.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameIntent, rs.Primary.ID, err)
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameIntent, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckIntentExists(ctx context.Context, name string, intent *lexmodelsv2.DescribeIntentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := conn.DescribeIntent(ctx, &lexmodelsv2.DescribeIntentInput{
			IntentId:   aws.String(rs.Primary.Attributes["intent_id"]),
			BotId:      aws.String(rs.Primary.Attributes["bot_id"]),
			BotVersion: aws.String(rs.Primary.Attributes["bot_version"]),
			LocaleId:   aws.String(rs.Primary.Attributes["locale_id"]),
		})

		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, rs.Primary.ID, err)
		}

		*intent = *resp

		return nil
	}
}

func testAccIntentConfig_base(rName string, ttl int, dp bool) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = %[1]q
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonLexFullAccess"
}

resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = %[2]d
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = %[3]t
  }
}

resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = "en_US"
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = 0.7
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
  locale_specification = {
    (aws_lexv2models_bot_locale.test.locale_id) = {
      source_bot_version = "DRAFT"
    }
  }
}
`, rName, ttl, dp)
}

func testAccIntentConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id
}
`, rName))
}

func testAccIntentConfig_updateConfirmationSetting(rName string, retries int, textMsg string, endTOMs1, endTOMs2 int) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  confirmation_setting {
    active = true

    prompt_specification {
      allow_interrupt            = true
      max_retries                = %[2]d
      message_selection_strategy = "Ordered"

      message_group {
        message {
          plain_text_message {
            value = %[3]q
          }
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Initial"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = %[4]d
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Retry1"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = %[5]d
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }
    }
  }
}
`, rName, retries, textMsg, endTOMs1, endTOMs2))
}

func testAccIntentConfig_updateClosingSetting(rName string, textMsg1, textMsg2, textMsg3 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  closing_setting {
    active = true

    closing_response {
      allow_interrupt = true

      message_group {
        message {
          plain_text_message {
            value = %[2]q
          }
        }
      }
    }

    conditional {
      active = true

      conditional_branch {
        name = %[1]q

        condition {
          expression_string = "slot1 = \"test\""
        }

        next_step {
          dialog_action {
            type           = "CloseIntent"
            slot_to_elicit = "slot1"
          }

          session_attributes = {
            slot1 = "roligt"
          }
        }

        response {
          allow_interrupt = true

          message_group {
            message {
              plain_text_message {
                value = %[3]q
              }
            }
          }
        }
      }

      default_branch {
        next_step {
          dialog_action {
            type           = "CloseIntent"
            slot_to_elicit = "slot1"
          }

          session_attributes = {
            slot1 = "hallon"
          }
        }

        response {
          allow_interrupt = true

          message_group {
            message {
              plain_text_message {
                value = %[4]q
              }
            }
          }
        }
      }
    }
  }
}
`, rName, textMsg1, textMsg2, textMsg3))
}

func testAccIntentConfig_updateInputContext(rName, inputContext1, inputContext2, inputContext3 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  input_context {
    name = %[2]q
  }

  input_context {
    name = %[3]q
  }

  input_context {
    name = %[4]q
  }
}
`, rName, inputContext1, inputContext2, inputContext3))
}

func testAccIntentConfig_updateInitialResponseSetting(rName, branchName, msg1, msg2, msg3 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  initial_response_setting {
    code_hook {
      active                      = true
      enable_code_hook_invocation = true
      invocation_label            = "test"

      post_code_hook_specification {
        failure_conditional {
          active = true

          conditional_branch {
            name = %[2]q

            condition {
              expression_string = "slot1 = \"test\""
            }

            next_step {
              dialog_action {
                type           = "CloseIntent"
                slot_to_elicit = "slot1"
              }

              session_attributes = {
                slot1 = "roligt"
              }
            }

            response {
              allow_interrupt = true

              message_group {
                message {
                  plain_text_message {
                    value = %[3]q
                  }
                }
              }
            }
          }

          default_branch {
            next_step {
              dialog_action {
                type           = "CloseIntent"
                slot_to_elicit = "slot1"
              }

              session_attributes = {
                slot1 = "hallon"
              }
            }

            response {
              allow_interrupt = true

              message_group {
                message {
                  plain_text_message {
                    value = %[4]q
                  }
                }
              }
            }
          }
        }
      }
    }

    initial_response {
      allow_interrupt = true

      message_group {
        message {
          plain_text_message {
            value = %[5]q
          }
        }
      }
    }
  }
}
`, rName, branchName, msg1, msg2, msg3))
}

func testAccIntentConfig_updateFulfillmentCodeHook(rName, msg1 string, freqSec int, msg2, condBranch, intIntValue, msg3 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  fulfillment_code_hook {
    active  = true
    enabled = true

    fulfillment_updates_specification {
      active             = true
      timeout_in_seconds = 10

      start_response {
        allow_interrupt  = true
        delay_in_seconds = 5

        message_group {
          message {
            plain_text_message {
              value = %[2]q
            }
          }
        }
      }

      update_response {
        allow_interrupt      = true
        frequency_in_seconds = %[3]d

        message_group {
          message {
            plain_text_message {
              value = %[4]q
            }
          }
        }
      }
    }

    post_fulfillment_status_specification {
      failure_conditional {
        active = true

        conditional_branch {
          name = %[5]q

          condition {
            expression_string = "slot1 = \"test\""
          }

          next_step {
            session_attributes = {
              slot1 = "roligt"
              slot2 = "roligt2"
            }

            dialog_action {
              type                  = "CloseIntent"
              slot_to_elicit        = "slot1"
              suppress_next_message = true
            }

            intent {
              name = "test"
              slot {
                map_block_key = "test"
                shape         = "Scalar"

                value {
                  interpreted_value = %[6]q
                }
              }
            }
          }

          response {
            allow_interrupt = true

            message_group {
              message {
                plain_text_message {
                  value = %[7]q
                }
              }
            }
          }
        }

        default_branch {
          next_step {
            dialog_action {
              type           = "CloseIntent"
              slot_to_elicit = "slot1"
            }

            session_attributes = {
              slot1 = "hallon"
            }
          }

          response {
            allow_interrupt = true

            message_group {
              message {
                plain_text_message {
                  value = "safriduo"
                }
              }
            }
          }
        }
      }
    }
  }
}
`, rName, msg1, freqSec, msg2, condBranch, intIntValue, msg3))
}

func testAccIntentConfig_updateDialogCodeHook(rName string, enabled bool) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  dialog_code_hook {
    enabled = %[2]t
  }
}
`, rName, enabled))
}

func testAccIntentConfig_updateOutputContext(rName, name1, name2, name3 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  output_context {
    name                    = %[2]q
    time_to_live_in_seconds = 300
    turns_to_live           = 5
  }

  output_context {
    name                    = %[3]q
    time_to_live_in_seconds = 300
    turns_to_live           = 5
  }

  output_context {
    name                    = %[4]q
    time_to_live_in_seconds = 300
    turns_to_live           = 5
  }
}
`, rName, name1, name2, name3))
}

func testAccIntentConfig_updateSampleUtterance(rName, utter1, utter2, utter3, utter4 string) string {
	return acctest.ConfigCompose(
		testAccIntentConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  sample_utterance {
    utterance = %[2]q
  }

  sample_utterance {
    utterance = %[3]q
  }

  sample_utterance {
    utterance = %[4]q
  }

  sample_utterance {
    utterance = %[5]q
  }
}
`, rName, utter1, utter2, utter3, utter4))
}
