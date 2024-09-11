// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexv2models_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lexmodelsv2"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/create"
	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexV2ModelsBotVersion_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var botversion lexmodelsv2.DescribeBotVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotVersionExists(ctx, resourceName, &botversion),
					resource.TestCheckResourceAttr(resourceName, "locale_specification.%", acctest.Ct1),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var botversion lexmodelsv2.DescribeBotVersionOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_version.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotVersionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotVersionConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotVersionExists(ctx, resourceName, &botversion),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceBotVersion, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotVersionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_version" {
				continue
			}

			_, err := tflexv2models.FindBotVersionByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameBotVersion, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBotVersionExists(ctx context.Context, name string, botversion *lexmodelsv2.DescribeBotVersionOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotVersion, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotVersion, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := tflexv2models.FindBotVersionByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotVersion, rs.Primary.ID, err)
		}

		*botversion = *resp

		return nil
	}
}

func testAccBotVersionConfig_basic(rName string) string {
	return acctest.ConfigCompose(
		testAccBotBaseConfig(rName),
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
