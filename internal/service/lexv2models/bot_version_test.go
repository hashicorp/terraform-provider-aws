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

func TestAccLexV2ModelsBotVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var botversion lexmodelsv2.DescribeBotVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotVersionExists(ctx, t, resourceName, &botversion),
					resource.TestCheckResourceAttr(resourceName, "locale_specification.%", "1"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_id"),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"locale_specification"},
			},
		},
	})
}

func TestAccLexV2ModelsBotVersion_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var botversion lexmodelsv2.DescribeBotVersionOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_version.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotVersionDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotVersionExists(ctx, t, resourceName, &botversion),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflexv2models.ResourceBotVersion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotVersionDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_version" {
				continue
			}

			_, err := tflexv2models.FindBotVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_version"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lex v2 Bot Version %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBotVersionExists(ctx context.Context, t *testing.T, n string, v *lexmodelsv2.DescribeBotVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		output, err := tflexv2models.FindBotVersionByTwoPartKey(ctx, conn, rs.Primary.Attributes["bot_id"], rs.Primary.Attributes["bot_version"])

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccBotVersionConfig_basic(rName string) string {
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

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
  locale_specification = {
    (aws_lexv2models_bot_locale.test.locale_id) = {
      source_bot_version = "DRAFT"
    }
  }
}
`, rName))
}
