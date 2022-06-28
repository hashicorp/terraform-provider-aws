package lexmodels_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func init() {
	acctest.RegisterServiceErrorCheckFunc(lexmodelbuildingservice.EndpointsID, testAccErrorCheckSkip)

}

func testAccErrorCheckSkip(t *testing.T) resource.ErrorCheckFunc {
	return acctest.ErrorCheckSkipMessagesContaining(t,
		"You can't set the enableModelImprovements field to false",
	)
}

func TestAccLexModelsBot_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					testAccCheckBotNotExists(testBotID, "1"),

					resource.TestCheckNoResourceAttr(rName, "abort_statement"),
					resource.TestCheckResourceAttrSet(rName, "arn"),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, "child_directed", "false"),
					resource.TestCheckNoResourceAttr(rName, "clarification_prompt"),
					resource.TestCheckResourceAttr(rName, "create_version", "false"),
					acctest.CheckResourceAttrRFC3339(rName, "created_date"),
					resource.TestCheckResourceAttr(rName, "description", "Bot to order flowers on the behalf of a user"),
					resource.TestCheckResourceAttr(rName, "detect_sentiment", "false"),
					resource.TestCheckResourceAttr(rName, "enable_model_improvements", "false"),
					resource.TestCheckResourceAttr(rName, "failure_reason", ""),
					resource.TestCheckResourceAttr(rName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckNoResourceAttr(rName, "intent"),
					acctest.CheckResourceAttrRFC3339(rName, "last_updated_date"),
					resource.TestCheckResourceAttr(rName, "locale", "en-US"),
					resource.TestCheckResourceAttr(rName, "name", testBotID),
					resource.TestCheckResourceAttr(rName, "nlu_intent_confidence_threshold", "0"),
					resource.TestCheckResourceAttr(rName, "process_behavior", "SAVE"),
					resource.TestCheckResourceAttr(rName, "status", "NOT_BUILT"),
					resource.TestCheckResourceAttr(rName, "version", tflexmodels.BotVersionLatest),
					resource.TestCheckNoResourceAttr(rName, "voice_id"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_Version_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"LexBot_createVersion":         testAccBot_createVersion,
		"LexBotAlias_botVersion":       testAccBotAlias_botVersion,
		"DataSourceLexBot_withVersion": testAccBotDataSource_withVersion,
		"DataSourceLexBotAlias_basic":  testAccBotAliasDataSource_basic,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccBot_createVersion(t *testing.T) {
	var v1, v2 lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	// If this test runs in parallel with other Lex Bot tests, it loses its description
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v1),
					testAccCheckBotNotExists(testBotID, "1"),
					resource.TestCheckResourceAttr(rName, "version", tflexmodels.BotVersionLatest),
					resource.TestCheckResourceAttr(rName, "description", "Bot to order flowers on the behalf of a user"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_createVersion(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(rName, "1", &v2),
					resource.TestCheckResourceAttr(rName, "version", "1"),
					resource.TestCheckResourceAttr(rName, "description", "Bot to order flowers on the behalf of a user"),
				),
			},
		},
	})
}

func TestAccLexModelsBot_abortStatement(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_abortStatement(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "abort_statement.#", "1"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.0.content", "Sorry, I'm not able to assist at this time"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckNoResourceAttr(rName, "abort_statement.0.message.0.group_number"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.response_card", ""),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_abortStatementUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.0.content", "Sorry, I'm not able to assist at this time"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.0.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.1.content", "Sorry, I'm not able to assist at this time. Good bye."),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.1.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.message.1.group_number", "1"),
					resource.TestCheckResourceAttr(rName, "abort_statement.0.response_card", "Sorry, I'm not able to assist at this time"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_clarificationPrompt(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_clarificationPrompt(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.#", "1"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.message.0.content", "I didn't understand you, what would you like to do?"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.message.0.content_type", "PlainText"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.response_card", ""),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_clarificationPromptUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.max_attempts", "3"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(rName, "clarification_prompt.0.response_card", "I didn't understand you, what would you like to do?"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_childDirected(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_childDirectedUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "child_directed", "true"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_description(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_descriptionUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "description", "Bot to order flowers"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_detectSentiment(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_detectSentimentUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "detect_sentiment", "true"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_enableModelImprovements(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_enableModelImprovementsUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "enable_model_improvements", "true"),
					resource.TestCheckResourceAttr(rName, "nlu_intent_confidence_threshold", "0.5"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_idleSessionTTLInSeconds(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_idleSessionTTLInSecondsUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "idle_session_ttl_in_seconds", "600"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_intents(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultiple(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultiple(testBotID),
					testAccBotConfig_intentsUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "intent.#", "2"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_computeVersion(t *testing.T) {
	var v1 lexmodelbuildingservice.GetBotOutput
	var v2 lexmodelbuildingservice.GetBotAliasOutput

	botResourceName := "aws_lex_bot.test"
	botAliasResourceName := "aws_lex_bot_alias.test"
	intentResourceName := "aws_lex_intent.test"
	intentResourceName2 := "aws_lex_intent.test_2"

	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	version := "1"
	updatedVersion := "2"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_createVersion(testBotID),
					testAccBotConfig_intentMultiple(testBotID),
					testAccBotAliasConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(botResourceName, version, &v1),
					resource.TestCheckResourceAttr(botResourceName, "version", version),
					resource.TestCheckResourceAttr(botResourceName, "intent.#", "1"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", version),
					testAccCheckBotAliasExists(botAliasResourceName, &v2),
					resource.TestCheckResourceAttr(botAliasResourceName, "bot_version", version),
					resource.TestCheckResourceAttr(intentResourceName, "version", version),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intentMultipleSecondUpdated(testBotID),
					testAccBotConfig_multipleIntentsWithVersion(testBotID),
					testAccBotAliasConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExistsWithVersion(botResourceName, updatedVersion, &v1),
					resource.TestCheckResourceAttr(botResourceName, "version", updatedVersion),
					resource.TestCheckResourceAttr(botResourceName, "intent.#", "2"),
					resource.TestCheckResourceAttr(botResourceName, "intent.0.intent_version", version),
					resource.TestCheckResourceAttr(botResourceName, "intent.1.intent_version", updatedVersion),
					resource.TestCheckResourceAttr(botAliasResourceName, "bot_version", updatedVersion),
					resource.TestCheckResourceAttr(intentResourceName, "version", version),
					resource.TestCheckResourceAttr(intentResourceName2, "version", updatedVersion),
				),
			},
		},
	})
}

func TestAccLexModelsBot_locale(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_localeUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "locale", "en-GB"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_voiceID(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_voiceIdUpdate(testBotID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "voice_id", "Justin"),
				),
			},
			{
				ResourceName:      rName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccLexModelsBot_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetBotOutput
	rName := "aws_lex_bot.test"
	testBotID := "test_bot_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:        acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckBotDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccBotConfig_intent(testBotID),
					testAccBotConfig_basic(testBotID),
				),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(rName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tflexmodels.ResourceBot(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotExistsWithVersion(rName, botVersion string, v *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex Bot ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn

		output, err := tflexmodels.FindBotVersionByName(conn, rs.Primary.ID, botVersion)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckBotExists(rName string, output *lexmodelbuildingservice.GetBotOutput) resource.TestCheckFunc {
	return testAccCheckBotExistsWithVersion(rName, tflexmodels.BotVersionLatest, output)
}

func testAccCheckBotNotExists(botName, botVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn

		_, err := tflexmodels.FindBotVersionByName(conn, botName, botVersion)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Lex Box %s/%s still exists", botName, botVersion)
	}
}

func testAccCheckBotDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lex_bot" {
			continue
		}

		output, err := conn.GetBotVersions(&lexmodelbuildingservice.GetBotVersionsInput{
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
