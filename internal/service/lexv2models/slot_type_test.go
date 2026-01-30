// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexV2ModelsSlotType_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var slottype lexmodelsv2.DescribeSlotTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot_type.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, t, resourceName, &slottype),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
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

func TestAccLexV2ModelsSlotType_values(t *testing.T) {
	ctx := acctest.Context(t)

	var slottype lexmodelsv2.DescribeSlotTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot_type.test"
	botLocaleName := "aws_lexv2models_bot_locale.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_values(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, t, resourceName, &slottype),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttrPair(resourceName, "bot_id", botLocaleName, "bot_id"),
					resource.TestCheckResourceAttrPair(resourceName, "bot_version", botLocaleName, "bot_version"),
					resource.TestCheckResourceAttrPair(resourceName, "locale_id", botLocaleName, "locale_id"),
					resource.TestCheckResourceAttrSet(resourceName, "slot_type_values.#"),
					resource.TestCheckResourceAttrSet(resourceName, "slot_type_values.0.%"),
					resource.TestCheckResourceAttr(resourceName, "slot_type_values.0.sample_value.0.value", "testval"),
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

func TestAccLexV2ModelsSlotType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var slottype lexmodelsv2.DescribeSlotTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, t, resourceName, &slottype),
					acctest.CheckFrameworkResourceDisappears(ctx, t, tflexv2models.ResourceSlotType, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexV2ModelsSlotType_valueSelectionSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var slottype lexmodelsv2.DescribeSlotTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_valueSelectionSetting(rName, string(types.AudioRecognitionStrategyUseSlotValuesAsCustomVocabulary)),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, t, resourceName, &slottype),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "value_selection_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "value_selection_setting.0.advanced_recognition_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "value_selection_setting.0.advanced_recognition_setting.0.audio_recognition_strategy", string(types.AudioRecognitionStrategyUseSlotValuesAsCustomVocabulary)),
				),
			},
		},
	})
}

func TestAccLexV2ModelsSlotType_compositeSlotTypeSetting(t *testing.T) {
	ctx := acctest.Context(t)

	var slottype lexmodelsv2.DescribeSlotTypeOutput
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_slot_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_compositeSlotTypeSetting(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, t, resourceName, &slottype),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "composite_slot_type_setting.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_slot_type_setting.0.sub_slots.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "composite_slot_type_setting.0.sub_slots.0.name", "testname"),
					resource.TestCheckResourceAttr(resourceName, "composite_slot_type_setting.0.sub_slots.0.slot_type_id", "AMAZON.Date"),
				),
			},
		},
	})
}

func testAccCheckSlotTypeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_slot" {
				continue
			}

			_, err := tflexv2models.FindSlotTypeByID(ctx, conn, rs.Primary.ID)
			if errs.IsA[*types.ResourceNotFoundException](err) {
				return nil
			}
			if err != nil {
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameSlotType, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckSlotTypeExists(ctx context.Context, t *testing.T, name string, slottype *lexmodelsv2.DescribeSlotTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlotType, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlotType, name, errors.New("not set"))
		}

		conn := acctest.ProviderMeta(ctx, t).LexV2ModelsClient(ctx)

		out, err := tflexv2models.FindSlotTypeByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameSlotType, rs.Primary.ID, err)
		}

		*slottype = *out

		return nil
	}
}

func testAccSlotTypeConfig_base(rName string, ttl int, dp bool) string {
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
`, rName, ttl, dp)
}

func testAccSlotTypeConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccSlotTypeConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot_type" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_selection_setting {
    resolution_strategy = "OriginalValue"
  }
}
`, rName))
}

func testAccSlotTypeConfig_values(rName string) string {
	return acctest.ConfigCompose(
		testAccSlotTypeConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot_type" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_selection_setting {
    resolution_strategy = "OriginalValue"
  }

  slot_type_values {
    sample_value {
      value = "testval"
    }
  }
  slot_type_values {
    sample_value {
      value = "testval2"
    }
  }
}
`, rName))
}

func testAccSlotTypeConfig_valueSelectionSetting(rName, audioRecognitionStrategy string) string {
	return acctest.ConfigCompose(
		testAccSlotTypeConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot_type" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_selection_setting {
    resolution_strategy = "OriginalValue"

    advanced_recognition_setting {
      audio_recognition_strategy = %[2]q
    }
  }

  slot_type_values {
    sample_value {
      value = "testval"
    }
  }
}
`, rName, audioRecognitionStrategy))
}

func testAccSlotTypeConfig_compositeSlotTypeSetting(rName string) string {
	return acctest.ConfigCompose(
		testAccSlotTypeConfig_base(rName, 60, true),
		fmt.Sprintf(`
resource "aws_lexv2models_slot_type" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = %[1]q
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  value_selection_setting {
    resolution_strategy = "Concatenation"
  }

  composite_slot_type_setting {
    sub_slots {
      name         = "testname"
      slot_type_id = "AMAZON.Date"
    }
  }
}
`, rName))
}
