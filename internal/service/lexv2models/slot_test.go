// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexV2ModelsSlot_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var slot lexmodelsv2.DescribeSlotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotExists(ctx, resourceName, &slot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttrSet(resourceName, "intent_id"),
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

func TestAccLexV2ModelsSlot_updateMultipleValuesSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var slot lexmodelsv2.DescribeSlotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotConfig_updateMultipleValuesSetting(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotExists(ctx, resourceName, &slot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.0.allow_multiple_values", acctest.CtTrue),
				),
			},
			{
				Config: testAccSlotConfig_updateMultipleValuesSetting(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotExists(ctx, resourceName, &slot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "multiple_values_setting.0.allow_multiple_values", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccLexV2ModelsSlot_ObfuscationSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var slot lexmodelsv2.DescribeSlotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotConfig_updateObfuscationSetting(rName, "DefaultObfuscation"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotExists(ctx, resourceName, &slot),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttr(resourceName, "obfuscation_setting.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "obfuscation_setting.0.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "obfuscation_setting.0.obfuscation_setting_type", "DefaultObfuscation"),
				),
			},
		},
	})
}

func TestAccLexV2ModelsSlot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var slot lexmodelsv2.DescribeSlotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotExists(ctx, resourceName, &slot),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceSlot, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckSlotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_slot" {
				continue
			}

			_, err := tflexv2models.FindSlotByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameSlot, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSlotExists(ctx context.Context, name string, slot *lexmodelsv2.DescribeSlotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlot, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlot, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		out, err := tflexv2models.FindSlotByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlot, rs.Primary.ID, err)
		}

		*slot = *out

		return nil
	}
}

func testAccSlotConfig_base(rName string, ttl int, dp bool) string {
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

resource "aws_lexv2models_intent" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id
}
`, rName, ttl, dp)
}

func testAccSlotConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSlotConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  intent_id   = aws_lexv2models_intent.test.intent_id
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_elicitation_setting {
    slot_constraint = "Optional"
    default_value_specification {
      default_value_list {
        default_value = "default"
      }
    }
  }
}
`, rName))
}

func testAccSlotConfig_updateMultipleValuesSetting(rName string, allow bool) string {
	return acctest.ConfigCompose(
		testAccSlotConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  intent_id   = aws_lexv2models_intent.test.intent_id
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_elicitation_setting {
    slot_constraint = "Optional"
    default_value_specification {
      default_value_list {
        default_value = "default"
      }
    }
  }

  multiple_values_setting {
    allow_multiple_values = %[2]t
  }
}
`, rName, allow))
}

func testAccSlotConfig_updateObfuscationSetting(rName, settingType string) string {
	return acctest.ConfigCompose(
		testAccSlotConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  intent_id   = aws_lexv2models_intent.test.intent_id
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_elicitation_setting {
    slot_constraint = "Optional"
    default_value_specification {
      default_value_list {
        default_value = "default"
      }
    }
  }

  obfuscation_setting {
    obfuscation_setting_type = %[2]q
  }
}
`, rName, settingType))
}
