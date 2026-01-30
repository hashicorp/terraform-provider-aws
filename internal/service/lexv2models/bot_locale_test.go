// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexV2ModelsBotLocale_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var botlocale lexmodelsv2.DescribeBotLocaleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_locale.test"
	botResourceName := "aws_lexv2models_bot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotLocaleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotLocaleConfig_basic(rName, "en_US", 0.7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotLocaleExists(ctx, t, resourceName, &botlocale),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botResourceName, names.AttrID),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "DRAFT"),
					resource.TestCheckResourceAttr(resourceName, "locale_id", "en_US"),
					resource.TestCheckResourceAttr(resourceName, "n_lu_intent_confidence_threshold", "0.7"),
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

func TestAccLexV2ModelsBotLocale_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var botlocale lexmodelsv2.DescribeBotLocaleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_locale.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotLocaleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotLocaleConfig_basic(rName, "en_US", 0.70),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotLocaleExists(ctx, t, resourceName, &botlocale),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflexv2models.ResourceBotLocale, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsBotLocale_voiceSettings(t *testing.T) {
	ctx := acctest.Context(t)
	var botlocale lexmodelsv2.DescribeBotLocaleOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_locale.test"
	// https://docs.aws.amazon.com/polly/latest/dg/voicelist.html
	voiceID := "Kendra"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotLocaleDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotLocaleConfig_voiceSettings(rName, voiceID, string(types.VoiceEngineStandard)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotLocaleExists(ctx, t, resourceName, &botlocale),
					resource.TestCheckTypeSetElemNestedAttrs(resourceName, "voice_settings.*", map[string]string{
						"voice_id":       voiceID,
						names.AttrEngine: string(types.VoiceEngineStandard),
					}),
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

func testAccCheckBotLocaleDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_locale" {
				continue
			}

			_, err := tflexv2models.FindBotLocaleByThreePartKey(ctx, conn, rs.Primary.Attributes["locale_id"], rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_version"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lex v2 Bot Locale %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBotLocaleExists(ctx context.Context, t *testing.T, n string, v *lexmodelsv2.DescribeBotLocaleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		output, err := tflexv2models.FindBotLocaleByThreePartKey(ctx, conn, rs.Primary.Attributes["locale_id"], rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_version"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBotLocaleConfig_base(rName string) string {
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
}`, rName))
}

func testAccBotLocaleConfig_basic(rName, localeID string, thres float64) string {
	return acctest.ConfigCompose(
		testAccBotLocaleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = %[1]q
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = %[2]g
}
`, localeID, thres))
}

func testAccBotLocaleConfig_voiceSettings(rName, voiceID, engine string) string {
	return acctest.ConfigCompose(
		testAccBotLocaleConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = "en_US"
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = 0.7

  voice_settings {
    voice_id = %[1]q
    engine   = %[2]q
  }
}
`, voiceID, engine))
}
