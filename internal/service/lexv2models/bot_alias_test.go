// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexV2ModelsBotAlias_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var botalias lexmodelsv2.DescribeBotAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, "bot_alias_name", rName),
					resource.TestCheckResourceAttrSet(resourceName, "bot_alias_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
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

func TestAccLexV2ModelsBotAlias_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var botalias lexmodelsv2.DescribeBotAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflexv2models.ResourceBotAlias, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsBotAlias_description(t *testing.T) {
	ctx := acctest.Context(t)
	var botalias lexmodelsv2.DescribeBotAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_description(rName, "first"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "first"),
				),
			},
			{
				Config: testAccBotAliasConfig_description(rName, "second"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, names.AttrDescription, "second"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsBotAlias_conversationLogSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var botalias lexmodelsv2.DescribeBotAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_conversationLogSettings(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, "conversation_log_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conversation_log_settings.0.text_log_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "conversation_log_settings.0.text_log_settings.0.enabled", "true"),
					resource.TestCheckResourceAttrSet(resourceName, "conversation_log_settings.0.text_log_settings.0.destination.0.cloudwatch.0.cloudwatch_log_group_arn"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsBotAlias_sentimentAnalysisSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var botalias lexmodelsv2.DescribeBotAliasOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_alias.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAliasDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAliasConfig_sentimentAnalysisSettings(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, "sentiment_analysis_settings.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "sentiment_analysis_settings.0.detect_sentiment", "true"),
				),
			},
			{
				Config: testAccBotAliasConfig_sentimentAnalysisSettings(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAliasExists(ctx, t, resourceName, &botalias),
					resource.TestCheckResourceAttr(resourceName, "sentiment_analysis_settings.0.detect_sentiment", "false"),
				),
			},
		},
	})
}

func testAccCheckBotAliasDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_alias" {
				continue
			}

			_, err := tflexv2models.FindBotAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_alias_id"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lex v2 Bot Alias %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBotAliasExists(ctx context.Context, t *testing.T, n string, v *lexmodelsv2.DescribeBotAliasOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		output, err := tflexv2models.FindBotAliasByTwoPartKey(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_alias_id"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBotAliasConfig_base(rName string) string {
	return acctest.ConfigCompose(
		testAccBotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = "true"
  }
}

resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = "en_US"
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = 0.7
}
`, rName))
}

func testAccBotAliasConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccBotAliasConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot_alias" "test" {
  bot_id         = aws_lexv2models_bot.test.id
  bot_alias_name = %[1]q

  depends_on = [aws_lexv2models_bot_locale.test]
}
`, rName))
}

func testAccBotAliasConfig_description(rName, description string) string {
	return acctest.ConfigCompose(
		testAccBotAliasConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot_alias" "test" {
  bot_id         = aws_lexv2models_bot.test.id
  bot_alias_name = %[1]q
  description    = %[2]q

  depends_on = [aws_lexv2models_bot_locale.test]
}
`, rName, description))
}

func testAccBotAliasConfig_conversationLogSettings(rName string) string {
	return acctest.ConfigCompose(
		testAccBotAliasConfig_base(rName),
		fmt.Sprintf(`
resource "aws_cloudwatch_log_group" "test" {
  name = %[1]q
}

resource "aws_lexv2models_bot_alias" "test" {
  bot_id         = aws_lexv2models_bot.test.id
  bot_alias_name = %[1]q

  conversation_log_settings {
    text_log_settings {
      enabled = true

      destination {
        cloudwatch {
          cloudwatch_log_group_arn = aws_cloudwatch_log_group.test.arn
          log_prefix               = "lex/"
        }
      }
    }
  }

  depends_on = [aws_lexv2models_bot_locale.test]
}
`, rName))
}

func testAccBotAliasConfig_sentimentAnalysisSettings(rName string, detect bool) string {
	return acctest.ConfigCompose(
		testAccBotAliasConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot_alias" "test" {
  bot_id         = aws_lexv2models_bot.test.id
  bot_alias_name = %[1]q

  sentiment_analysis_settings {
    detect_sentiment = %[2]t
  }

  depends_on = [aws_lexv2models_bot_locale.test]
}
`, rName, detect))
}
