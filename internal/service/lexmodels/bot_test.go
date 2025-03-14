// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(names.LexModelsServiceID, testAccErrorCheckSkip)
}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"You can't set the enableModelImprovements field to false",
	)
}

func TestAccLexModelsBot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					testAccCheckBotNotExists(ctx, testBotName, "1"),

					resource.TestCheckResourceAttr(resourceName, "abort_statement.#", "1"),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "lex", "bot:{name}"),
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttr(resourceName, "child_directed", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "create_version", acctest.CtFalse),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrCreatedDate),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Bot to order flowers on the behalf of a user"),
					resource.TestCheckResourceAttr(resourceName, "detect_sentiment", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "enable_model_improvements", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "failure_reason", ""),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "intent.#", "1"),
					acctest.CheckResourceAttrRFC3339(resourceName, names.AttrLastUpdatedDate),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-US"),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, testBotName),
					resource.TestCheckResourceAttr(resourceName, "nlu_intent_confidence_threshold", "0"),
					resource.TestCheckResourceAttr(resourceName, "process_behavior", "SAVE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrStatus, "NOT_BUILT"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttr(resourceName, "voice_id", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_Version_serial(t *testing.T) {
	t.Parallel()

	testCases := map[string]func(t *testing.T){
		"LexBot_createVersion":         testAccBot_createVersion,
		"LexBotAlias_botVersion":       testAccBotAlias_botVersion,
		"DataSourceLexBot_withVersion": testAccBotDataSource_withVersion,
		"DataSourceLexBotAlias_basic":  testAccBotAliasDataSource_basic,
	}

	acctest.RunSerialTests1Level(t, testCases, 0)
}

func testAccBot_createVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1, v2 lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v1),
					testAccCheckBotNotExists(ctx, testBotName, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Bot to order flowers on the behalf of a user"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_createVersion(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(ctx, resourceName, "1", &v2),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Bot to order flowers on the behalf of a user"),
				),
			},
		},
	})
}

func TestAccLexModelsBot_abortStatement(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_abortStatement(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.0.content", "Sorry, I'm not able to assist at this time"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckNoResourceAttr(resourceName, "abort_statement.0.message.0.group_number"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.response_card", ""),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_abortStatementUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.0.content", "Sorry, I'm not able to assist at this time"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.0.group_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.1.content", "Sorry, I'm not able to assist at this time. Good bye."),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.1.group_number", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.response_card", "Sorry, I'm not able to assist at this time"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_clarificationPrompt(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_clarificationPrompt(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.0.content", "I didn't understand you, what would you like to do?"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.response_card", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"abort_statement.0.message.0.group_number",
					"clarification_prompt.0.message.0.group_number",
				},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_clarificationPromptUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.max_attempts", "3"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.response_card", "I didn't understand you, what would you like to do?"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"abort_statement.0.message.0.group_number",
					"clarification_prompt.0.message.0.group_number",
				},
			},
		},
	})
}

func TestAccLexModelsBot_childDirected(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_childDirectedUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "child_directed", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_descriptionUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "Bot to order flowers"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_detectSentiment(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_detectSentimentUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "detect_sentiment", acctest.CtTrue),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_enableModelImprovements(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_enableModelImprovementsUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "enable_model_improvements", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "nlu_intent_confidence_threshold", "0.5"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_idleSessionTTLInSeconds(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_idleSessionTTLInSecondsUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "600"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_intents(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultiple(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultiple(testBotName),
					testAccBotConfig_intentsUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "intent.#", "2"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_computeVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 lexmodelbuildingservice.GetBotOutput
	var v2 lexmodelbuildingservice.GetBotAliasOutput

	botResourceName := "aws_lex_bot.test"
	botAliasResourceName := "aws_lex_bot_alias.test"
	intentResourceName := "aws_lex_intent.test"
	intentResourceName2 := "aws_lex_intent.test_2"

	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_createVersion(testBotName),
					testAccBotConfig_intentMultiple(testBotName),
					testAccBotAliasConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(ctx, botResourceName, "1", &v1),
					resource.TestCheckResourceAttr(botResourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(botResourceName, "intent.#", "1"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", "1"),
					testAccCheckBotAliasExists(ctx, botAliasResourceName, &v2),
					resource.TestCheckResourceAttr(botAliasResourceName, "bot_version", "1"),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, "1"),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultipleSecondUpdated(testBotName),
					testAccBotConfig_multipleIntentsWithVersion(testBotName),
					testAccBotAliasConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(ctx, botResourceName, "2", &v1),
					resource.TestCheckResourceAttr(botResourceName, names.AttrVersion, "2"),
					resource.TestCheckResourceAttr(botResourceName, "intent.#", "2"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", "1"),
					resource.TestCheckResourceAttr(botResourceName, "intent.1.intent_version", "2"),
					resource.TestCheckResourceAttr(botAliasResourceName, "bot_version", "2"),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, "1"),
					resource.TestCheckResourceAttr(intentResourceName2, names.AttrVersion, "2"),
				),
			},
		},
	})
}

func TestAccLexModelsBot_locale(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_localeUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-GB"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_voiceID(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_voiceIdUpdate(testBotName),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					resource.TestCheckResourceAttr(resourceName, "voice_id", "Justin"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"abort_statement.0.message.0.group_number"},
			},
		},
	})
}

func TestAccLexModelsBot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetBotOutput
	resourceName := "aws_lex_bot.test"
	testBotName := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexModelBuildingServiceEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotName),
					testAccBotConfig_basic(testBotName),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflexmodels.ResourceBot(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotExistsWithVersion(ctx context.Context, resourceName, botVersion string, v *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex Bot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		output, err := tflexmodels.FindBotVersionByName(ctx, conn, rs.Primary.ID, botVersion)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBotExists(ctx context.Context, resourceName string, output *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return testAccCheckBotExistsWithVersion(ctx, resourceName, tflexmodels.BotVersionLatest, output)
}

func testAccCheckBotNotExists(ctx context.Context, botName, botVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		_, err := tflexmodels.FindBotVersionByName(ctx, conn, botName, botVersion)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Lex Box %s/%s still exists", botName, botVersion)
	}
}

func testAccCheckBotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lex_bot" {
				continue
			}

			output, err := conn.GetBotVersions(ctx, &lexmodelbuildingservice.GetBotVersionsInput{
				Name: aws.String(rs.Primary.ID),
			})

			if err != nil {
				return err
			}

			if output == nil || len(output.Bots) == 0 {
				return nil
			}

			return fmt.Errorf("Lex Bot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccBotConfig_intent(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}
`, rName)
}

func testAccBotConfig_intentMultiple(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%[1]s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}

resource "aws_lex_intent" "test_2" {
  create_version = true
  name           = "%[1]stwo"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}
`, rName)
}

func testAccBotConfig_intentMultipleSecondUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = "%[1]s"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}

resource "aws_lex_intent" "test_2" {
  create_version = true
  name           = "%[1]stwo"
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to return these flowers",
  ]
}
`, rName)
}

func testAccBotConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_createVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed   = false
  create_version   = true
  description      = "Bot to order flowers on the behalf of a user"
  name             = "%s"
  process_behavior = "BUILD"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_abortStatement(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_abortStatementUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
      group_number = 1
    }
    message {
      content      = "Sorry, I'm not able to assist at this time. Good bye."
      content_type = "PlainText"
      group_number = 1
    }
    response_card = "Sorry, I'm not able to assist at this time"
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_clarificationPrompt(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  clarification_prompt {
    max_attempts = 2
    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_clarificationPromptUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  clarification_prompt {
    max_attempts = 3
    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
      group_number = 1
    }
    message {
      content      = "I didn't understand you, can you re-phrase your request, please?"
      content_type = "PlainText"
      group_number = 1
    }
    response_card = "I didn't understand you, what would you like to do?"
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_childDirectedUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = true
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_descriptionUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_detectSentimentUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed   = false
  description      = "Bot to order flowers on the behalf of a user"
  detect_sentiment = true
  name             = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_enableModelImprovementsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed                  = false
  description                     = "Bot to order flowers on the behalf of a user"
  enable_model_improvements       = true
  name                            = "%s"
  nlu_intent_confidence_threshold = 0.5
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_idleSessionTTLInSecondsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed              = false
  description                 = "Bot to order flowers on the behalf of a user"
  idle_session_ttl_in_seconds = 600
  name                        = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_intentsUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
  intent {
    intent_name    = aws_lex_intent.test_2.name
    intent_version = aws_lex_intent.test_2.version
  }
}
`, rName)
}

func testAccBotConfig_localeUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed            = false
  description               = "Bot to order flowers on the behalf of a user"
  enable_model_improvements = true
  locale                    = "en-GB"
  name                      = "%s"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_voiceIdUpdate(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed = false
  description    = "Bot to order flowers on the behalf of a user"
  name           = "%s"
  voice_id       = "Justin"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
}
`, rName)
}

func testAccBotConfig_multipleIntentsWithVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_bot" "test" {
  child_directed   = false
  create_version   = true
  description      = "Bot to order flowers on the behalf of a user"
  name             = "%s"
  process_behavior = "BUILD"
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
      content_type = "PlainText"
    }
  }
  intent {
    intent_name    = aws_lex_intent.test.name
    intent_version = aws_lex_intent.test.version
  }
  intent {
    intent_name    = aws_lex_intent.test_2.name
    intent_version = aws_lex_intent.test_2.version
  }
}
`, rName)
}
