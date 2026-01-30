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

func TestAccLexV2ModelsBot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var bot lexmodelsv2.DescribeBotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"
	iamRoleResourceName := "aws_iam_role.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_basic(rName, 60, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "60"),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrRoleARN, iamRoleResourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "data_privacy.0.child_directed"),
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

func TestAccLexV2ModelsBot_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var bot lexmodelsv2.DescribeBotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_tags1(rName, 60, true, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccBotConfig_tags2(rName, 60, true, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccBotConfig_tags1(rName, 60, true, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccLexV2ModelsBot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var bot lexmodelsv2.DescribeBotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_basic(rName, 60, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflexv2models.ResourceBot, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsBot_type(t *testing.T) {
	ctx := acctest.Context(t)
	var bot lexmodelsv2.DescribeBotOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_type(rName, 60, true, string(types.BotTypeBot)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, t, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, names.AttrType, string(types.BotTypeBot)),
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

func testAccCheckBotDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot" {
				continue
			}

			_, err := tflexv2models.FindBotByID(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Lex v2 Bot %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBotExists(ctx context.Context, t *testing.T, n string, v *lexmodelsv2.DescribeBotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		output, err := tflexv2models.FindBotByID(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

	input := &lexmodelsv2.ListBotsInput{}
	_, err := conn.ListBots(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

func testAccBotConfig_base(rName string) string {
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
`, rName)
}

func testAccBotConfig_basic(rName string, ttl int, dp bool) string {
	return acctest.ConfigCompose(
		testAccBotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = %[2]d
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = %[3]t
  }
}
`, rName, ttl, dp))
}

func testAccBotConfig_tags1(rName string, ttl int, dp bool, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccBotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = %[2]d
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = %[3]t
  }

  tags = {
    %[4]q = %[5]q
  }
}
`, rName, ttl, dp, tagKey1, tagValue1))
}

func testAccBotConfig_tags2(rName string, ttl int, dp bool, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccBotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = %[2]d
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = %[3]t
  }

  tags = {
    %[4]q = %[5]q
    %[6]q = %[7]q
  }
}
`, rName, ttl, dp, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccBotConfig_type(rName string, ttl int, dp bool, botType string) string {
	return acctest.ConfigCompose(
		testAccBotConfig_base(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = %[2]d
  role_arn                    = aws_iam_role.test.arn
  type                        = %[4]q

  data_privacy {
    child_directed = %[3]t
  }
}
`, rName, ttl, dp, botType))
}
