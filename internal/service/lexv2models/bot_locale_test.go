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

func TestAccLexV2ModelsBotLocale_basic(t *testing.T) {
	ctx := acctest.Context(t)

	var botlocale lexmodelsv2.DescribeBotLocaleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotLocaleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotLocaleConfig_basic(rName, "en_US", 0.7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotLocaleExists(ctx, resourceName, &botlocale),
					resource.TestCheckResourceAttrSet(resourceName, "bot_id"),
					resource.TestCheckResourceAttrSet(resourceName, "locale_id"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_version"),
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
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var botlocale lexmodelsv2.DescribeBotLocaleOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot_locale.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotLocaleDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotLocaleConfig_basic(rName, "en_US", 0.70),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotLocaleExists(ctx, resourceName, &botlocale),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceBotLocale, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotLocaleDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot_locale" {
				continue
			}

			_, err := tflexv2models.FindBotLocaleByID(ctx, conn, rs.Primary.ID)
			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameBotLocale, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBotLocaleExists(ctx context.Context, name string, botlocale *lexmodelsv2.DescribeBotLocaleOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotLocale, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBotLocale, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := tflexv2models.FindBotLocaleByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.AuditManager, create.ErrActionCheckingExistence, tflexv2models.ResNameBotLocale, rs.Primary.ID, err)
		}

		*botlocale = *resp

		return nil
	}
}

func testAccBotLocaleConfig_basic(rName, localeid string, thres float64) string {
	return acctest.ConfigCompose(
		testAccBotBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                        = %[1]q
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.test_role.arn

  data_privacy {
    child_directed = "true"
  }
}

resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = %[2]q
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = %[3]g
}
`, rName, localeid, thres))
}
