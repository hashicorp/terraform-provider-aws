// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

// **PLEASE DELETE THIS AND ALL TIP COMMENTS BEFORE SUBMITTING A PR FOR REVIEW!**
//
// TIP: ==== INTRODUCTION ====
// Thank you for trying the skaff tool!
//
// You have opted to include these helpful comments. They all include "TIP:"
// to help you find and remove them when you're done with them.
//
// While some aspects of this file are customized to your input, the
// scaffold tool does *not* look at the AWS API and ensure it has correct
// function, structure, and variable names. It makes guesses based on
// commonalities. You will need to make significant adjustments.
//
// In other words, as generated, this is a rough outline of the work you will
// need to do. If something doesn't make sense for your situation, get rid of
// it.

import (
	// TIP: ==== IMPORTS ====
	// This is a common set of imports but not customized to your code since
	// your code hasn't been written yet. Make sure you, your IDE, or
	// goimports -w <file> fixes these imports.
	//
	// The provider linter wants your imports to be in two groups: first,
	// standard library (i.e., "fmt" or "strings"), second, everything else.
	//
	// Also, AWS Go SDK v2 may handle nested structures differently than v1,
	// using the services/lexv2models/types package. If so, you'll
	// need to import types and reference the nested types, e.g., as
	// types.<Type Name>.
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"

	"github.com/aws/aws-sdk-go-v2/aws"
	lextypes "github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-provider-aws/internal/framework/flex"
	fwtypes "github.com/hashicorp/terraform-provider-aws/internal/framework/types"

	// TIP: You will often need to import the package that this test file lives
	// in. Since it is in the "test" context, it must import the package to use
	// any normal context constants, variables, or functions.
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
)

// TIP: File Structure. The basic outline for all test files should be as
// follows. Improve this resource's maintainability by following this
// outline.
//
// 1. Package declaration (add "_test" since this is a test file)
// 2. Imports
// 3. Unit tests
// 4. Basic test
// 5. Disappears test
// 6. All the other tests
// 7. Helper functions (exists, destroy, check, etc.)
// 8. Functions that return Terraform configurations

// TIP: ==== UNIT TESTS ====
// This is an example of a unit test. Its name is not prefixed with
// "TestAcc" like an acceptance test.
//
// Unlike acceptance tests, unit tests do not access AWS and are focused on a
// function (or method). Because of this, they are quick and cheap to run.
//
// In designing a resource's implementation, isolate complex bits from AWS bits
// so that they can be tested through a unit test. We encourage more unit tests
// in the provider.
//
// Cut and dry functions using well-used patterns, like typical flatteners and
// expanders, don't need unit testing. However, if they are complex or
// intricate, they should be unit tested.

// TestIntentAutoExpand tests autoflex expand for the specific intent data
// structures. As such, this is not meant to be a comprehensive test of autoflex
// generally, but a test to make sure the complex intent data structures work
// with autoflex now and in the future.
func TestIntentAutoExpand(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	testString := "b72d06fd-2b78-5fe2-a6a3-e06e5efde347"
	//testString2 := "a47c2004-f58b-5982-880a-f68c80f6307c"

	ssmlMessage := tflexv2models.SSMLMessage{
		Value: types.StringValue(testString),
	}
	plainTextMessage := tflexv2models.PlainTextMessage{
		Value: types.StringValue(testString),
	}
	button := tflexv2models.Button{
		Text:  types.StringValue(testString),
		Value: types.StringValue(testString),
	}
	buttons := []tflexv2models.Button{
		button,
	}
	imageResponseCard := tflexv2models.ImageResponseCard{
		Title:    types.StringValue(testString),
		Button:   fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.Button](ctx, buttons),
		ImageURL: types.StringValue(testString),
		Subtitle: types.StringValue(testString),
	}
	customPayload := tflexv2models.CustomPayload{
		Value: types.StringValue(testString),
	}
	message := tflexv2models.Message{
		CustomPayload:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &customPayload),
		ImageResponseCard: fwtypes.NewListNestedObjectValueOfPtr(ctx, &imageResponseCard),
		PlainTextMessage:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &plainTextMessage),
		SSMLMessage:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &ssmlMessage),
	}
	messages := []tflexv2models.Message{
		message,
	}
	messageGroup := tflexv2models.MessageGroup{
		Message:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &message),
		Variations: fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.Message](ctx, messages),
	}
	responseSpecification := tflexv2models.ResponseSpecification{
		MessageGroup:   fwtypes.NewListNestedObjectValueOfPtr[tflexv2models.MessageGroup](ctx, &messageGroup),
		AllowInterrupt: types.BoolValue(true),
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
	intentOverride := tflexv2models.IntentOverride{
		Name: types.StringValue(testString),
		//Slots: fwtypes.NewMapValueOf[tflexv2models.SlotValueOverride](ctx, &slotValueOverride),
	}
	dialogAction := tflexv2models.DialogAction{
		Type:                types.StringValue(testString),
		SlotToElicit:        types.StringValue(testString),
		SuppressNextMessage: types.BoolValue(true),
	}
	dialogState := tflexv2models.DialogState{
		DialogAction: fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogAction),
		Intent:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentOverride),
		/*SessionAttributes: types.MapValueMust(types.StringType, map[string]attr.Value{
			testString: types.StringValue(testString2),
		}),*/
	}
	condition := tflexv2models.Condition{
		ExpressionString: types.StringValue(testString),
	}
	defaultConditionalBranch := tflexv2models.DefaultConditionalBranch{
		NextStep: fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		Response: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
	}
	conditionalBranch := tflexv2models.ConditionalBranch{
		Condition: fwtypes.NewListNestedObjectValueOfPtr(ctx, &condition),
		Name:      types.StringValue(testString),
		NextStep:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		Response:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
	}
	conditionalBranches := []tflexv2models.ConditionalBranch{
		conditionalBranch,
	}
	conditionalSpecification := tflexv2models.ConditionalSpecification{
		Active:            types.BoolValue(true),
		ConditionalBranch: fwtypes.NewListNestedObjectValueOfValueSlice(ctx, conditionalBranches),
		DefaultBranch:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &defaultConditionalBranch),
	}
	intentClosingSetting := tflexv2models.IntentClosingSetting{
		Active:          types.BoolValue(true),
		ClosingResponse: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		Conditional:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		NextStep:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
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
		}*/
	promptSpecification := tflexv2models.PromptSpecification{
		MaxRetries:               types.Int64Value(1),
		MessageGroup:             fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroup),
		AllowInterrupt:           types.BoolValue(true),
		MessageSelectionStrategy: types.StringValue(testString),
		//PromptAttemptsSpecification: promptAttemptSpecificationMap,
	}
	failureSuccessTimeout := tflexv2models.FailureSuccessTimeout{
		FailureConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		FailureNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		FailureResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		SuccessConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		SuccessNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		SuccessResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		TimeoutConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		TimeoutNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		TimeoutResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
	}
	dialogCodeHookInvocationSetting := tflexv2models.DialogCodeHookInvocationSetting{
		Active:                    types.BoolValue(true),
		EnableCodeHookInvocation:  types.BoolValue(true),
		InvocationLabel:           types.StringValue(testString),
		PostCodeHookSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &failureSuccessTimeout),
	}
	elicitationCodeHook := tflexv2models.ElicitationCodeHookInvocationSetting{
		EnableCodeHookInvocation: types.BoolValue(true),
		InvocationLabel:          types.StringValue(testString),
	}
	intentConfirmationSetting := tflexv2models.IntentConfirmationSetting{
		PromptSpecification:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &promptSpecification),
		Active:                  types.BoolValue(true),
		CodeHook:                fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookInvocationSetting),
		ConfirmationConditional: fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		ConfirmationNextStep:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		ConfirmationResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		DeclinationConditional:  fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		DeclinationNextStep:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		DeclinationResponse:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		ElicitationCodeHook:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &elicitationCodeHook),
		FailureConditional:      fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		FailureNextStep:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
		FailureResponse:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
	}
	dialogCodeHookSettings := tflexv2models.DialogCodeHookSettings{
		Enabled: types.BoolValue(true),
	}
	fulfillmentStartResponseSpecification := tflexv2models.FulfillmentStartResponseSpecification{
		DelayInSeconds: types.Int64Value(1),
		MessageGroup:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroup),
		AllowInterrupt: types.BoolValue(true),
	}
	fulfillmentUpdateResponseSpecification := tflexv2models.FulfillmentUpdateResponseSpecification{
		FrequencyInSeconds: types.Int64Value(1),
		MessageGroup:       fwtypes.NewListNestedObjectValueOfPtr(ctx, &messageGroup),
		AllowInterrupt:     types.BoolValue(true),
	}
	fulfillmentUpdatesSpecification := tflexv2models.FulfillmentUpdatesSpecification{
		Active:           types.BoolValue(true),
		StartResponse:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentStartResponseSpecification),
		TimeoutInSeconds: types.Int64Value(1),
		UpdateResponse:   fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentUpdateResponseSpecification),
	}
	fulfillmentCodeHookSettings := tflexv2models.FulfillmentCodeHookSettings{
		Enabled:                            types.BoolValue(true),
		Active:                             types.BoolValue(true),
		FulfillmentUpdatesSpecification:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentUpdatesSpecification),
		PostFulfillmentStatusSpecification: fwtypes.NewListNestedObjectValueOfPtr(ctx, &failureSuccessTimeout),
	}
	initialResponseSetting := tflexv2models.InitialResponseSetting{
		CodeHook:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookInvocationSetting),
		Conditional:     fwtypes.NewListNestedObjectValueOfPtr(ctx, &conditionalSpecification),
		InitialResponse: fwtypes.NewListNestedObjectValueOfPtr(ctx, &responseSpecification),
		NextStep:        fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogState),
	}
	inputContext := tflexv2models.InputContext{
		Name: types.StringValue(testString),
	}
	inputContexts := []tflexv2models.InputContext{
		inputContext,
	}
	kendraConfiguration := tflexv2models.KendraConfiguration{
		KendraIndex:              types.StringValue(testString),
		QueryFilterString:        types.StringValue(testString),
		QueryFilterStringEnabled: types.BoolValue(true),
	}
	outputContext := tflexv2models.OutputContext{
		Name:                types.StringValue(testString),
		TimeToLiveInSeconds: types.Int64Value(1),
		TurnsToLive:         types.Int64Value(1),
	}
	outputContexts := []tflexv2models.OutputContext{
		outputContext,
	}
	sampleUtterance := tflexv2models.SampleUtterance{
		Utterance: types.StringValue(testString),
	}
	sampleUtterances := []tflexv2models.SampleUtterance{
		sampleUtterance,
	}
	slotPriority := tflexv2models.SlotPriority{
		Priority: types.Int64Value(1),
		SlotID:   types.StringValue(testString),
	}
	slotPriorities := []tflexv2models.SlotPriority{
		slotPriority,
	}
	intentResource := tflexv2models.ResourceIntentData{
		BotID:                  types.StringValue(testString),
		BotVersion:             types.StringValue(testString),
		ClosingSetting:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentClosingSetting),
		ConfirmationSetting:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &intentConfirmationSetting),
		CreationDateTime:       fwtypes.TimestampType{},
		Description:            types.StringValue(testString),
		DialogCodeHook:         fwtypes.NewListNestedObjectValueOfPtr(ctx, &dialogCodeHookSettings),
		FulfillmentCodeHook:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &fulfillmentCodeHookSettings),
		ID:                     types.StringValue(testString),
		InitialResponseSetting: fwtypes.NewListNestedObjectValueOfPtr(ctx, &initialResponseSetting),
		InputContext:           fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.InputContext](ctx, inputContexts),
		KendraConfiguration:    fwtypes.NewListNestedObjectValueOfPtr(ctx, &kendraConfiguration),
		LastUpdatedDateTime:    fwtypes.TimestampType{},
		LocaleID:               types.StringValue(testString),
		Name:                   types.StringValue(testString),
		OutputContext:          fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.OutputContext](ctx, outputContexts),
		ParentIntentSignature:  types.StringValue(testString),
		SampleUtterance:        fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.SampleUtterance](ctx, sampleUtterances),
		SlotPriority:           fwtypes.NewListNestedObjectValueOfValueSlice[tflexv2models.SlotPriority](ctx, slotPriorities),
	}
	fmt.Printf("intentResource %s\n", intentResource)

	testCases := []struct {
		TestName   string
		Source     any
		Target     any
		WantErr    bool
		WantTarget any
	}{
		{
			TestName: "message",
			Source:   &message,
			Target:   &lextypes.Message{},
			WantTarget: &lextypes.Message{
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
			},
		},
		{
			TestName: "plainTextMessage",
			Source:   &plainTextMessage,
			Target:   &lextypes.PlainTextMessage{},
			WantTarget: &lextypes.PlainTextMessage{
				Value: aws.String(testString),
			},
		},
		{
			TestName: "button",
			Source:   &button,
			Target:   &lextypes.Button{},
			WantTarget: &lextypes.Button{
				Text:  aws.String(testString),
				Value: aws.String(testString),
			},
		},
		/*
			{
				TestName: "complex Source and complex Target",
				Source:   &intentResource,
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
