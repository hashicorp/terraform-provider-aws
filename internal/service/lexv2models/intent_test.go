// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/aws"
	lextypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
)

func TestIntentAutoExpand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testString := "b72d06fd-2b78-5fe2-a6a3-e06e5efde347"
	testString2 := "a47c2004-f58b-5982-880a-f68c80f6307c"

	ssmlMessageTF := tflexv2models.SSMLMessage{
		Value: types.StringValue(testString),
	}
	plainTextMessageTF := tflexv2models.PlainTextMessage{
		Value: types.StringValue(testString),
	}
	buttonTF := tflexv2models.Button{
		Text:  types.StringValue(testString),
		Value: types.StringValue(testString),
	}
	buttonsTF := []tflexv2models.Button{
		buttonTF,
	}
	imageResponseCardTF := tflexv2models.ImageResponseCard{
		Title:    types.StringValue(testString),
		Button:   fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.Button](ctx, buttonsTF),
		ImageURL: types.StringValue(testString),
		Subtitle: types.StringValue(testString),
	}
	customPayloadTF := tflexv2models.CustomPayload{
		Value: types.StringValue(testString),
	}
	messageTF := tflexv2models.Message{
		CustomPayload:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &customPayloadTF),
		ImageResponseCard: fwtypes.NewListNestedObjectValueOfPtr(ctx, &imageResponseCardTF),
		PlainTextMessage:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &plainTextMessageTF),
		SSMLMessage:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &ssmlMessageTF),
	}
	messageAWS := lextypes.Message{
		CustomPayload: &lextypes.CustomPayload{
			Value: aws.String(testString),
		},
		ImageResponseCard: &lextypes.ImageResponseCard{
			Title: aws.String(testString),
			Buttons: []lextypes.Button{{
				Text:  aws.String(testString),
				Value: aws.String(testString),
			}},
			ImageUrl: aws.String(testString),
			Subtitle: aws.String(testString),
		},
		PlainTextMessage: &lextypes.PlainTextMessage{
			Value: aws.String(testString),
		},
		SsmlMessage: &lextypes.SSMLMessage{
			Value: aws.String(testString),
		},
	}
	messagesTF := []tflexv2models.Message{
		messageTF,
	}
	messageGroupTF := tflexv2models.MessageGroup{
		Message:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageTF),
		Variations: fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.Message](ctx, messagesTF),
	}
	responseSpecificationTF := tflexv2models.ResponseSpecification{
		MessageGroup:   fwtypes.NewListNestedObjectValueOfPtr[tflexv2models.MessageGroup](ctx, &messageGroupTF),
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
	/*
		slotValue := tflexv2models.SlotValue{
			InterpretedValue: types.StringValue(testString),
		}
	*/
	/*
		slotValueOverride := tflexv2models.SlotValueOverride{
			Shape: types.StringValue(testString),
			Value: fwtypes.NewListNestedObjectValueOfPtr(ctx, &slotValue),
			Values: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []tflexv2models.SlotValueOverride{ // recursive so must be defined in line instead of in variable
				{
					Shape: types.StringValue(testString),
					Value: fwtypes.NewListNestedObjectValueOfPtr(ctx, &slotValue),
				},
			}),
		}
	*/
	intentOverrideTF := tflexv2models.IntentOverride{
		Name: types.StringValue(testString),
		//Slots: fwtypes.NewMapValueOf[tflexv2models.SlotValueOverride](ctx, &slotValueOverride),
	}
	dialogActionTF := tflexv2models.DialogAction{
		Type:                types.StringValue(string(lextypes.DialogActionTypeCloseIntent)),
		SlotToElicit:        types.StringValue(testString),
		SuppressNextMessage: types.BoolValue(true),
	}
	dialogStateTF := tflexv2models.DialogState{
		DialogAction: fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogActionTF),
		Intent:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentOverrideTF),
		SessionAttributes: types.MapValueMust(types.StringType, map[string]attr.Value{
			testString: types.StringValue(testString2),
		}),
	}
	dialogStateAWS := lextypes.DialogState{
		DialogAction: &lextypes.DialogAction{
			Type:                lextypes.DialogActionTypeCloseIntent,
			SlotToElicit:        aws.String(testString),
			SuppressNextMessage: aws.Bool(true),
		},
		Intent: &lextypes.IntentOverride{
			Name: aws.String(testString),
		},
		SessionAttributes: map[string]string{
			testString: testString2,
		},
	}

	//conditionTF := tflexv2models.Condition{
	//	ExpressionString: types.StringValue(testString),
	//}

	defaultConditionalBranchTF := tflexv2models.DefaultConditionalBranch{
		NextStep: fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
		Response: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
	}

	/*
		conditionalBranchTF := tflexv2models.ConditionalBranch{
			Condition: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionTF),
			Name:      types.StringValue(testString),
			//NextStep:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			Response: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
		}
	*/

	//conditionalBranchesTF := []tflexv2models.ConditionalBranch{
	//	conditionalBranchTF,
	//}
	conditionalSpecificationTF := tflexv2models.ConditionalSpecification{
		Active: types.BoolValue(true),
		/*
			ConditionalBranch: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, []tflexv2models.ConditionalBranch{
				{
					Condition: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionTF),
					Name: types.StringValue(testString),
					NextStep:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
					Response: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
				},
			}),
		*/
		DefaultBranch: fwtypes.NewListNestedObjectValueOfPtr(ctx, &defaultConditionalBranchTF),
	}
	conditionalSpecificationAWS := lextypes.ConditionalSpecification{
		Active: aws.Bool(true),
		ConditionalBranches: []lextypes.ConditionalBranch{
			{
				Condition: &lextypes.Condition{
					ExpressionString: aws.String(testString),
				},
				Name: aws.String(testString),
				//NextStep: &dialogStateAWS,
				Response: &responseSpecificationAWS,
			},
		},
		DefaultBranch: &lextypes.DefaultConditionalBranch{
			//NextStep: &dialogStateAWS,
			Response: &responseSpecificationAWS,
		},
	}
	/*
		intentClosingSettingTF := tflexv2models.IntentClosingSetting{
			Active:          types.BoolValue(true),
			ClosingResponse: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			Conditional:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			NextStep:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
		}
		intentClosingSettingAWS := lextypes.IntentClosingSetting{
			Active:          aws.Bool(true),
			ClosingResponse: &responseSpecificationAWS,
			Conditional:     &conditionalSpecificationAWS,
		}

		/*
				allowedInputTypes := tflexv2models.AllowedInputTypes{
					AllowAudioInput: types.BoolValue(true),
					AllowDTMFInput:  types.BoolValue(true),
				}
				audioSpecification := tflexv2models.AudioSpecification{
					EndTimeoutMs: types.Int64Value(1),
					MaxLengthMs:  types.Int64Value(1),
				}
				dtmfSpecification := tflexv2models.DTMFSpecification{
					DeletionCharacter: types.StringValue(testString),
					EndCharacter:      types.StringValue(testString),
					EndTimeoutMs:      types.Int64Value(1),
					MaxLength:         types.Int64Value(1),
				}
				audioAndDTMFInputSpecification := tflexv2models.AudioAndDTMFInputSpecification{
					StartTimeoutMs:     types.Int64Value(1),
					AudioSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &audioSpecification),
					DTMFSpecification:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &dtmfSpecification),
				}

				textInputSpecification := tflexv2models.TextInputSpecification{
					StartTimeoutMs: types.Int64Value(1),
				}
				promptAttemptSpecification := tflexv2models.PromptAttemptSpecification{
					AllowedInputTypes:              fwtypes.NewListNestedObjectValueOfPtr(ctx, &allowedInputTypes),
					AllowInterrupt:                 types.BoolValue(true),
					AudioAndDTMFInputSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &audioAndDTMFInputSpecification),
					TextInputSpecification:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &textInputSpecification),
				}
				promptAttemptSpecificationMap := map[string]tflexv2models.PromptAttemptSpecification{
					testString: promptAttemptSpecification,
				}
			promptSpecificationTF := tflexv2models.PromptSpecification{
				MaxRetries:               types.Int64Value(1),
				MessageGroup:             fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroupTF),
				AllowInterrupt:           types.BoolValue(true),
				MessageSelectionStrategy: types.StringValue(testString),
				//PromptAttemptsSpecification: promptAttemptSpecificationMap,
			}
	*/
	/*
		failureSuccessTimeoutTF := tflexv2models.FailureSuccessTimeout{
			FailureConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			FailureNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			FailureResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			SuccessConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			SuccessNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			SuccessResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			TimeoutConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			TimeoutNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			TimeoutResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
		}

		dialogCodeHookInvocationSettingTF := tflexv2models.DialogCodeHookInvocationSetting{
			Active:                    types.BoolValue(true),
			EnableCodeHookInvocation:  types.BoolValue(true),
			InvocationLabel:           types.StringValue(testString),
			PostCodeHookSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &failureSuccessTimeoutTF),
		}
	*/
	/*
		elicitationCodeHookTF := tflexv2models.ElicitationCodeHookInvocationSetting{
			EnableCodeHookInvocation: types.BoolValue(true),
			InvocationLabel:          types.StringValue(testString),
		}
		/*
		intentConfirmationSettingTF := tflexv2models.IntentConfirmationSetting{
			PromptSpecification:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &promptSpecificationTF),
			Active:                  types.BoolValue(true),
			CodeHook:                fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookInvocationSettingTF),
			ConfirmationConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			ConfirmationNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			ConfirmationResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			DeclinationConditional:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			DeclinationNextStep:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			DeclinationResponse:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			ElicitationCodeHook:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &elicitationCodeHookTF),
			FailureConditional:      fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			FailureNextStep:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
			FailureResponse:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
		}
	*/
	dialogCodeHookSettingsTF := tflexv2models.DialogCodeHookSettings{
		Enabled: types.BoolValue(true),
	}
	/*
		fulfillmentStartResponseSpecificationTF := tflexv2models.FulfillmentStartResponseSpecification{
			DelayInSeconds: types.Int64Value(1),
			MessageGroup:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroupTF),
			AllowInterrupt: types.BoolValue(true),
		}
		fulfillmentUpdateResponseSpecificationTF := tflexv2models.FulfillmentUpdateResponseSpecification{
			FrequencyInSeconds: types.Int64Value(1),
			MessageGroup:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroupTF),
			AllowInterrupt:     types.BoolValue(true),
		}
		fulfillmentUpdatesSpecificationTF := tflexv2models.FulfillmentUpdatesSpecification{
			Active:           types.BoolValue(true),
			StartResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentStartResponseSpecificationTF),
			TimeoutInSeconds: types.Int64Value(1),
			UpdateResponse:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentUpdateResponseSpecificationTF),
		}

		fulfillmentCodeHookSettingsTF := tflexv2models.FulfillmentCodeHookSettings{
			Enabled:                            types.BoolValue(true),
			Active:                             types.BoolValue(true),
			FulfillmentUpdatesSpecification:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentUpdatesSpecificationTF),
			PostFulfillmentStatusSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &failureSuccessTimeoutTF),
		}
		initialResponseSettingTF := tflexv2models.InitialResponseSetting{
			CodeHook:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookInvocationSettingTF),
			Conditional:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecificationTF),
			InitialResponse: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecificationTF),
			NextStep:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogStateTF),
		}
	*/
	inputContextTF := tflexv2models.InputContext{
		Name: types.StringValue(testString),
	}
	inputContextsTF := []tflexv2models.InputContext{
		inputContextTF,
	}
	kendraConfigurationTF := tflexv2models.KendraConfiguration{
		KendraIndex:              types.StringValue(testString),
		QueryFilterString:        types.StringValue(testString),
		QueryFilterStringEnabled: types.BoolValue(true),
	}
	outputContextTF := tflexv2models.OutputContext{
		Name:                types.StringValue(testString),
		TimeToLiveInSeconds: types.Int64Value(1),
		TurnsToLive:         types.Int64Value(1),
	}
	outputContextsTF := []tflexv2models.OutputContext{
		outputContextTF,
	}
	sampleUtteranceTF := tflexv2models.SampleUtterance{
		Utterance: types.StringValue(testString),
	}
	sampleUtterancesTF := []tflexv2models.SampleUtterance{
		sampleUtteranceTF,
	}
	slotPriorityTF := tflexv2models.SlotPriority{
		Priority: types.Int64Value(1),
		SlotID:   types.StringValue(testString),
	}
	slotPrioritiesTF := []tflexv2models.SlotPriority{
		slotPriorityTF,
	}
	intentResourceTF := tflexv2models.ResourceIntentData{
		BotID:      types.StringValue(testString),
		BotVersion: types.StringValue(testString),
		//ClosingSetting:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentClosingSettingTF),
		//ConfirmationSetting:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentConfirmationSettingTF),
		CreationDateTime: fwtypes.TimestampType{},
		Description:      types.StringValue(testString),
		DialogCodeHook:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookSettingsTF),
		//FulfillmentCodeHook:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentCodeHookSettingsTF),
		ID: types.StringValue(testString),
		//InitialResponseSetting: fwtypes.NewListNestedObjectValueOfPtr(ctx, &initialResponseSettingTF),
		InputContext:          fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.InputContext](ctx, inputContextsTF),
		KendraConfiguration:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &kendraConfigurationTF),
		LastUpdatedDateTime:   fwtypes.TimestampType{},
		LocaleID:              types.StringValue(testString),
		Name:                  types.StringValue(testString),
		OutputContext:         fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.OutputContext](ctx, outputContextsTF),
		ParentIntentSignature: types.StringValue(testString),
		SampleUtterance:       fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.SampleUtterance](ctx, sampleUtterancesTF),
		SlotPriority:          fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.SlotPriority](ctx, slotPrioritiesTF),
	}
	fmt.Printf("intentResourceTF %s\n", intentResourceTF)

	testCases := []struct {
		TestName   string
		Source     any
		Target     any
		WantErr    bool
		WantTarget any
	}{
		{
			TestName:   "message",
			Source:     &messageTF,
			Target:     &lextypes.Message{},
			WantTarget: &messageAWS,
		},
		{
			TestName:   "responseSpecification",
			Source:     &responseSpecificationTF,
			Target:     &lextypes.ResponseSpecification{},
			WantTarget: &responseSpecificationAWS,
		},
		{
			TestName:   "dialogState",
			Source:     &dialogStateTF,
			Target:     &lextypes.DialogState{},
			WantTarget: &dialogStateAWS,
		},
		{
			TestName:   "conditionalSpecification",
			Source:     &conditionalSpecificationTF,
			Target:     &lextypes.ConditionalSpecification{},
			WantTarget: &conditionalSpecificationAWS,
		},

		/*
			{
				TestName: "complex Source and complex Target",
				Source:   &intentResourceTF,
				Target:   &lexmodelsv2.CreateIntentInput{},
				WantTarget: &lexmodelsv2.CreateIntentInput{
					BotId: aws.String(testString),
				},
			},
		*/
	}

	opts := cmpopts.IgnoreUnexported(
		lextypes.SSMLMessage{},
		lextypes.PlainTextMessage{},
		lextypes.Button{},
		lextypes.CustomPayload{},
		lextypes.ImageResponseCard{},
		lextypes.Message{},
		lextypes.ResponseSpecification{},
		lextypes.MessageGroup{},
		lextypes.DialogAction{},
		lextypes.DialogState{},
		lextypes.IntentOverride{},
		lextypes.ConditionalSpecification{},
		lextypes.ConditionalBranch{},
	)

	for _, testCase := range testCases {
		testCase := testCase
		t.Run(testCase.TestName, func(t *testing.T) {
			t.Parallel()

			err := flex.Expand(ctx, testCase.Source, testCase.Target)
			gotErr := err != nil

			if gotErr != testCase.WantErr {
				t.Errorf("gotErr = %v, wantErr = %v", gotErr, testCase.WantErr)
			}

			if gotErr {
				if !testCase.WantErr {
					t.Errorf("err = %q", err)
				}
			} else if diff := cmp.Diff(testCase.Target, testCase.WantTarget, opts); diff != "" {
				t.Errorf("unexpected diff (+wanted, -got): %s", diff)
			}
		})
	}
}

/*
// TIP: ==== ACCEPTANCE TESTS ====
// This is an example of a basic acceptance test. This should test as much of
// standard functionality of the resource as possible, and test importing, if
// applicable. We prefix its name with "TestAcc", the service, and the
// resource name.
//
// Acceptance test access AWS and cost money to run.
func TestAccLexV2ModelsIntent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var intent lexv2models.DescribeIntentResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					resource.TestCheckResourceAttr(resourceName, "auto_minor_version_upgrade", "false"),
					resource.TestCheckResourceAttrSet(resourceName, "maintenance_window_start_time.0.day_of_week"),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "user.*", map[string]string{
						"console_access": "false",
						"groups.#":       "0",
						"username":       "Test",
						"password":       "TestTest1234",
					}),
					acctest.MatchResourceAttrRegionalARN(resourceName, "arn", "lexv2models", regexache.MustCompile(`intent:+.`)),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"apply_immediately", "user"},
			},
		},
	})
}

func TestAccLexV2ModelsIntent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var intent lexv2models.DescribeIntentResponse
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_intent.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(rName, testAccIntentVersionNewer),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, resourceName, &intent),
					// TIP: The Plugin-Framework disappears helper is similar to the Plugin-SDK version,
					// but expects a new resource factory function as the third argument. To expose this
					// private function to the testing package, you may need to add a line like the following
					// to exports_test.go:
					//
					//   var ResourceIntent = newResourceIntent
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceIntent, resourceName),
				),
				ExpectNonEmptyPlan: true,
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

			input := &lexv2models.DescribeIntentInput{
				IntentId: aws.String(rs.Primary.ID),
			}
			_, err := conn.DescribeIntent(ctx, &lexv2models.DescribeIntentInput{
				IntentId: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*types.ResourceNotFoundException](err) {
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

func testAccCheckIntentExists(ctx context.Context, name string, intent *lexv2models.DescribeIntentResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := conn.DescribeIntent(ctx, &lexv2models.DescribeIntentInput{
			IntentId: aws.String(rs.Primary.ID),
		})

		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameIntent, rs.Primary.ID, err)
		}

		*intent = *resp

		return nil
	}
}

func testAccCheckIntentNotRecreated(before, after *lexv2models.DescribeIntentResponse) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if before, after := aws.ToString(before.IntentId), aws.ToString(after.IntentId); before != after {
			return create.Error(names.LexV2Models, create.ErrActionCheckingNotRecreated, tflexv2models.ResNameIntent, aws.ToString(before.IntentId), errors.New("recreated"))
		}

		return nil
	}
}

func testAccIntentConfig_basic(rName, version string) string {
	return fmt.Sprintf(`
resource "aws_security_group" "test" {
  name = %[1]q
}

resource "aws_lexv2models_intent" "test" {
  intent_name             = %[1]q
  engine_type             = "ActiveLexV2Models"
  engine_version          = %[2]q
  host_instance_type      = "lexv2models.t2.micro"
  security_groups         = [aws_security_group.test.id]
  authentication_strategy = "simple"
  storage_type            = "efs"

  logs {
    general = true
  }

  user {
    username = "Test"
    password = "TestTest1234"
  }
}
`, rName, version)
}
*/
