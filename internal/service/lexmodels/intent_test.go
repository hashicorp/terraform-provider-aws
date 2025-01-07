// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	awstypes "github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice/types"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/retry"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexModelsIntent_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					testAccCheckIntentNotExists(ctx, testIntentID, "1"),

					resource.TestCheckResourceAttrSet(rName, names.AttrARN),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.#", "0"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.#", "0"),
					resource.TestCheckResourceAttr(rName, "create_version", acctest.CtFalse),
					acctest.CheckResourceAttrRFC3339(rName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(rName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(rName, "dialog_code_hook.#", "0"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.#", "0"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.#", "1"),
					acctest.CheckResourceAttrRFC3339(rName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttr(rName, names.AttrName, testIntentID),
					resource.TestCheckResourceAttr(rName, "parent_intent_signature.#", "0"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.#", "0"),
					resource.TestCheckNoResourceAttr(rName, "sample_utterances"),
					resource.TestCheckResourceAttr(rName, "slot.#", "0"),
					resource.TestCheckResourceAttr(rName, names.AttrVersion, tflexmodels.IntentVersionLatest),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccLexModelsIntent_createVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					testAccCheckIntentNotExists(ctx, testIntentID, "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccIntentConfig_createVersion(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					testAccCheckIntentExistsWithVersion(ctx, rName, "1", &v),
					resource.TestCheckResourceAttr(rName, names.AttrVersion, "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccLexModelsIntent_conclusionStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_conclusionStatement(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckNoResourceAttr(rName, "conclusion_statement.0.message.0.group_number"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.response_card", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"conclusion_statement.0.message.0.group_number",
				},
			},
			{
				Config: testAccIntentConfig_conclusionStatementUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content", "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.0.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.content", "Your order for {FlowerType} has been placed"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.message.1.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "conclusion_statement.0.response_card", "Your order for {FlowerType} has been placed"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"conclusion_statement.0.message.0.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsIntent_confirmationPromptAndRejectionStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_confirmationPromptAndRejectionStatement(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content", "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content", "Okay, I will not place your order."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.response_card", "Okay, I will not place your order."),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"confirmation_prompt.0.message.0.group_number",
					"rejection_statement.0.message.0.group_number",
				},
			},
			{
				Config: testAccIntentConfig_confirmationPromptAndRejectionStatementUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content", "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.1.content", "Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "confirmation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content", "Okay, I will not place your order."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.1.content", "Okay, your order has been cancelled."),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "rejection_statement.0.response_card", "Okay, your order has been cancelled."),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"confirmation_prompt.0.message.0.group_number",
					"confirmation_prompt.0.message.1.group_number",
					"rejection_statement.0.message.0.group_number",
					"rejection_statement.0.message.1.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsIntent_dialogCodeHook(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccIntentConfig_lambda(testIntentID),
					testAccIntentConfig_dialogCodeHook(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "dialog_code_hook.#", "1"),
					resource.TestCheckResourceAttr(rName, "dialog_code_hook.0.message_version", "1"),
					resource.TestCheckResourceAttrSet(rName, "dialog_code_hook.0.uri"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccLexModelsIntent_followUpPrompt(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_followUpPrompt(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.#", "1"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content", "Would you like to order more flowers?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to order more flowers?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content", "Okay, no additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.response_card", "Okay, no additional flowers will be ordered."),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"follow_up_prompt.0.prompt.0.message.0.group_number",
					"follow_up_prompt.0.rejection_statement.0.message.0.group_number",
				},
			},
			{
				Config: testAccIntentConfig_followUpPromptUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content", "Would you like to order more flowers?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.1.content", "Would you like to start another order?"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to start another order?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"),

					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content", "Okay, additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.1.content", "Okay, no additional flowers will be ordered."),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "follow_up_prompt.0.rejection_statement.0.response_card", "Okay, additional flowers will be ordered."),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"follow_up_prompt.0.prompt.0.message.0.group_number",
					"follow_up_prompt.0.prompt.0.message.1.group_number",
					"follow_up_prompt.0.rejection_statement.0.message.0.group_number",
					"follow_up_prompt.0.rejection_statement.0.message.1.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsIntent_fulfillmentActivity(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccIntentConfig_lambda(testIntentID),
					testAccIntentConfig_fulfillmentActivity(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.#", "1"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.code_hook.#", "1"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.code_hook.0.message_version", "1"),
					resource.TestCheckResourceAttrSet(rName, "fulfillment_activity.0.code_hook.0.uri"),
					resource.TestCheckResourceAttr(rName, "fulfillment_activity.0.type", "CodeHook"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccLexModelsIntent_sampleUtterances(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_sampleUtterances(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "sample_utterances.0", "I would like to pick up flowers"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccIntentConfig_sampleUtterancesUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "sample_utterances.#", "2"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccLexModelsIntent_slots(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_slots(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.description", "The date to pick up the flowers"),
					resource.TestCheckResourceAttr(rName, "slot.0.name", "PickupDate"),
					resource.TestCheckResourceAttr(rName, "slot.0.priority", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.0", "I would like to order {FlowerType}"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_constraint", "Required"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_type", "AMAZON.DATE"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.max_attempts", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content", "What day do you want the {FlowerType} to be picked up?"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content_type", "PlainText"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"slot.0.value_elicitation_prompt.0.message.0.group_number",
				},
			},
			{
				Config: testAccIntentConfig_slotsUpdate(testIntentID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "2"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"slot.0.value_elicitation_prompt.0.message.0.group_number",
					"slot.1.value_elicitation_prompt.0.message.0.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsIntent_slotsCustom(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_basic(testIntentID),
					testAccIntentConfig_slotsCustom(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "slot.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.description", "Types of flowers to pick up"),
					resource.TestCheckResourceAttr(rName, "slot.0.name", "FlowerType"),
					resource.TestCheckResourceAttr(rName, "slot.0.priority", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.sample_utterances.0", "I would like to order {FlowerType}"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_constraint", "Required"),
					resource.TestCheckResourceAttr(rName, "slot.0.slot_type", testIntentID),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content", "What type of flowers would you like to order?"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "slot.0.value_elicitation_prompt.0.response_card", "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"What type of flowers?\",\"buttons\":[{\"text\":\"Tulips\",\"value\":\"tulips\"},{\"text\":\"Lilies\",\"value\":\"lilies\"},{\"text\":\"Roses\",\"value\":\"roses\"}]}]}"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"create_version",
					"slot.0.value_elicitation_prompt.0.message.0.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsIntent_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflexmodels.ResourceIntent(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexModelsIntent_updateWithExternalChange(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetIntentOutput
	rName := "aws_lex_intent.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	testAccCheckAWSLexIntentUpdateDescription := func(provider *schema.Provider, _ *schema.Resource, resourceName string) resource.TestCheckFunc {
		return func(s *terraform.State) error {
			conn := provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

			resourceState, ok := s.RootModule().Resources[resourceName]
			if !ok {
				return fmt.Errorf("intent not found: %s", resourceName)
			}

			input := &lexmodelbuildingservice.PutIntentInput{
				Checksum:    aws.String(resourceState.Primary.Attributes["checksum"]),
				Description: aws.String("Updated externally without Terraform"),
				Name:        aws.String(resourceState.Primary.ID),
				FulfillmentActivity: &awstypes.FulfillmentActivity{
					Type: awstypes.FulfillmentActivityType("ReturnIntent"),
				},
			}
			err := retry.RetryContext(ctx, 1*time.Minute, func() *retry.RetryError {
				_, err := conn.PutIntent(ctx, input)

				if errs.IsA[*awstypes.ConflictException](err) {
					return retry.RetryableError(fmt.Errorf("%q: intent still updating", resourceName))
				}
				if err != nil {
					return retry.NonRetryableError(err)
				}

				return nil
			})
			if err != nil {
				return fmt.Errorf("error updating intent %s: %w", resourceName, err)
			}

			return nil
		}
	}

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
					testAccCheckAWSLexIntentUpdateDescription(acctest.Provider, tflexmodels.ResourceIntent(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccIntentConfig_basic(testIntentID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckIntentExists(ctx, rName, &v),
				),
			},
		},
	})
}

func TestAccLexModelsIntent_computeVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 lexmodelbuildingservice.GetIntentOutput
	var v2 lexmodelbuildingservice.GetBotOutput

	intentResourceName := "aws_lex_intent.test"
	botResourceName := "aws_lex_bot.test"
	testIntentID := "test_intent_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckIntentDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testIntentID),
					testAccBotConfig_createVersion(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExistsWithVersion(ctx, intentResourceName, "1", &v1),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(intentResourceName, "sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(intentResourceName, "sample_utterances.0", "I would like to pick up flowers"),
					testAccCheckBotExistsWithVersion(ctx, botResourceName, "1", &v2),
					resource.TestCheckResourceAttr(botResourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", "1"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccIntentConfig_sampleUtterancesWithVersion(testIntentID),
					testAccBotConfig_createVersion(testIntentID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckIntentExistsWithVersion(ctx, intentResourceName, "2", &v1),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(intentResourceName, "sample_utterances.#", "1"),
					resource.TestCheckResourceAttr(intentResourceName, "sample_utterances.0", "I would not like to pick up flowers"),
					testAccCheckBotExistsWithVersion(ctx, botResourceName, "2", &v2),
					resource.TestCheckResourceAttr(botResourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", "2"),
				),
			},
		},
	})
}

func testAccCheckIntentExistsWithVersion(ctx context.Context, rName, intentVersion string, output *lexmodelbuildingservice.GetIntentOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex intent ID is set")
		}

		var err error
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		output, err = conn.GetIntent(ctx, &lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(rs.Primary.ID),
			Version: aws.String(intentVersion),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return fmt.Errorf("error intent %q version %s not found", rs.Primary.ID, intentVersion)
		}
		if err != nil {
			return fmt.Errorf("error getting intent %q version %s: %w", rs.Primary.ID, intentVersion, err)
		}

		return nil
	}
}

func testAccCheckIntentExists(ctx context.Context, rName string, output *lexmodelbuildingservice.GetIntentOutput) resource.TestCheckFunc {
	return testAccCheckIntentExistsWithVersion(ctx, rName, tflexmodels.IntentVersionLatest, output)
}

func testAccCheckIntentNotExists(ctx context.Context, intentName, intentVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		_, err := conn.GetIntent(ctx, &lexmodelbuildingservice.GetIntentInput{
			Name:    aws.String(intentName),
			Version: aws.String(intentVersion),
		})
		if errs.IsA[*awstypes.NotFoundException](err) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting intent %s version %s: %s", intentName, intentVersion, err)
		}

		return fmt.Errorf("error intent %s version %s exists", intentName, intentVersion)
	}
}

func testAccCheckIntentDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lex_intent" {
				continue
			}

			output, err := conn.GetIntentVersions(ctx, &lexmodelbuildingservice.GetIntentVersionsInput{
				Name: aws.String(rs.Primary.ID),
			})
			if errs.IsA[*awstypes.NotFoundException](err) {
				continue
			}
			if err != nil {
				return err
			}

			if output == nil || len(output.Intents) == 0 {
				return nil
			}

			return fmt.Errorf("Lex intent %q still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccIntentConfig_lambda(rName string) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "lambda_assume_role" {
  statement {
    actions = ["sts:AssumeRole"]
    principals {
      type        = "Service"
      identifiers = ["lambda.amazonaws.com"]
    }
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.lambda_assume_role.json
  name               = "%[1]s"
}

resource "aws_lambda_permission" "lex" {
  action        = "lambda:InvokeFunction"
  function_name = aws_lambda_function.test.function_name
  principal     = "lex.amazonaws.com"
}

resource "aws_lambda_function" "test" {
  filename      = "test-fixtures/lambdatest.zip"
  function_name = "%[1]s"
  handler       = "lambdatest.handler"
  role          = aws_iam_role.test.arn
  runtime       = "nodejs20.x"
}
`, rName)
}

func testAccIntentConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
}
`, rName)
}

func testAccIntentConfig_createVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
}
`, rName)
}

func testAccIntentConfig_conclusionStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
    }
    response_card = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
  }
}
`, rName)
}

func testAccIntentConfig_conclusionStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  conclusion_statement {
    message {
      content      = "Your order for {FlowerType} has been placed and will be ready by {PickupTime} on {PickupDate}"
      content_type = "PlainText"
      group_number = 1
    }
    message {
      content      = "Your order for {FlowerType} has been placed"
      content_type = "PlainText"
      group_number = 1
    }
    response_card = "Your order for {FlowerType} has been placed"
  }
}
`, rName)
}

func testAccIntentConfig_confirmationPromptAndRejectionStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  confirmation_prompt {
    max_attempts = 1
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
  }
  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
    response_card = "Okay, I will not place your order."
  }
}
`, rName)
}

func testAccIntentConfig_confirmationPromptAndRejectionStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  confirmation_prompt {
    max_attempts = 2
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup by {PickupTime} on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    message {
      content      = "Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?"
      content_type = "PlainText"
    }
    response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Okay, your {FlowerType} will be ready for pickup on {PickupDate}. Does this sound okay?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
  }
  rejection_statement {
    message {
      content      = "Okay, I will not place your order."
      content_type = "PlainText"
    }
    message {
      content      = "Okay, your order has been cancelled."
      content_type = "PlainText"
    }
    response_card = "Okay, your order has been cancelled."
  }
}
`, rName)
}

func testAccIntentConfig_dialogCodeHook(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  dialog_code_hook {
    message_version = "1"
    uri             = aws_lambda_function.test.arn
  }
}
`, rName)
}

func testAccIntentConfig_followUpPrompt(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  follow_up_prompt {
    prompt {
      max_attempts = 1
      message {
        content      = "Would you like to order more flowers?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to order more flowers?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
    }
    rejection_statement {
      message {
        content      = "Okay, no additional flowers will be ordered."
        content_type = "PlainText"
      }
      response_card = "Okay, no additional flowers will be ordered."
    }
  }
}
`, rName)
}

func testAccIntentConfig_followUpPromptUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  follow_up_prompt {
    prompt {
      max_attempts = 2
      message {
        content      = "Would you like to order more flowers?"
        content_type = "PlainText"
      }
      message {
        content      = "Would you like to start another order?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"Would you like to start another order?\",\"buttons\":[{\"text\":\"Yes\",\"value\":\"yes\"},{\"text\":\"No\",\"value\":\"no\"}]}]}"
    }
    rejection_statement {
      message {
        content      = "Okay, no additional flowers will be ordered."
        content_type = "PlainText"
      }
      message {
        content      = "Okay, additional flowers will be ordered."
        content_type = "PlainText"
      }
      response_card = "Okay, additional flowers will be ordered."
    }
  }
}
`, rName)
}

func testAccIntentConfig_fulfillmentActivity(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    code_hook {
      message_version = "1"
      uri             = aws_lambda_function.test.arn
    }
    type = "CodeHook"
  }
}
`, rName)
}

func testAccIntentConfig_sampleUtterances(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}
`, rName)
}

func testAccIntentConfig_sampleUtterancesUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
    "I would like to order some flowers",
  ]
}
`, rName)
}

func testAccIntentConfig_slots(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"
    value_elicitation_prompt {
      max_attempts = 1
      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }
}
`, rName)
}

func testAccIntentConfig_slotsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "The date to pick up the flowers"
    name        = "PickupDate"
    priority    = 2
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.DATE"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "What day do you want the {FlowerType} to be picked up?"
        content_type = "PlainText"
      }
    }
  }
  slot {
    description = "The time to pick up the flowers"
    name        = "PickupTime"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint = "Required"
    slot_type       = "AMAZON.TIME"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "Pick up the {FlowerType} at what time on {PickupDate}?"
        content_type = "PlainText"
      }
    }
  }
}
`, rName)
}

func testAccIntentConfig_slotsCustom(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  name = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "Types of flowers to pick up"
    name        = "FlowerType"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint   = "Required"
    slot_type         = aws_lex_slot_type.test.name
    slot_type_version = "$LATEST"
    value_elicitation_prompt {
      max_attempts = 2
      message {
        content      = "What type of flowers would you like to order?"
        content_type = "PlainText"
      }
      response_card = "{\"version\":1,\"contentType\":\"application/vnd.amazonaws.card.generic\",\"genericAttachments\":[{\"title\":\"What type of flowers?\",\"buttons\":[{\"text\":\"Tulips\",\"value\":\"tulips\"},{\"text\":\"Lilies\",\"value\":\"lilies\"},{\"text\":\"Roses\",\"value\":\"roses\"}]}]}"
    }
  }
}
`, rName)
}

func testAccIntentConfig_sampleUtterancesWithVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would not like to pick up flowers",
  ]
}
`, rName)
}

func testAccIntentConfig_slotsWithVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  slot {
    description = "Types of flowers to pick up"
    name        = "FlowerType"
    priority    = 1
    sample_utterances = [
      "I would like to order {FlowerType}",
    ]
    slot_constraint   = "Required"
    slot_type         = aws_lex_slot_type.test.name
    slot_type_version = aws_lex_slot_type.test.version
  }
}
`, rName)
}
