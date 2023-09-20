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
	"github.com/hashicorp/terraform-provider-aws/names"

	tflexv2models "github.com/hashicorp/terraform-provider-aws/internal/service/lexv2models"
)

func TestAccLexV2ModelsBot_basic(t *testing.T) {
	ctx := acctest.Context(t)
	// TIP: This is a long-running test guard for tests that run longer than
	// 300s (5 min) generally.
	// if testing.Short() {
	// 	t.Skip("skipping long-running test in short mode")
	// }

	var bot lexmodelsv2.DescribeBotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &bot),
					resource.TestCheckResourceAttr(resourceName, "name", rName),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "5"),
					resource.TestCheckResourceAttr(resourceName, "role_arn", "bot_role_arn"),
					resource.TestCheckResourceAttr(resourceName, "data_privacy.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "data_privacy.0.child_directed", "true"),
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

func TestAccLexV2ModelsBot_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var bot lexmodelsv2.DescribeBotOutput
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_lexv2models_bot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.LexV2ModelsEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexV2ModelsEndpointID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotExists(ctx, resourceName, &bot),
					acctest.CheckFrameworkResourceDisappears(ctx, acctest.Provider, tflexv2models.ResourceBot, resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lexv2models_bot" {
				continue
			}

			_, err := tflexv2models.FindBotByID(ctx, conn, rs.Primary.ID)
			if err != nil {
				var nfe *types.ResourceNotFoundException
				if errors.As(err, &nfe) {
					return nil
				}
				return err
			}

			return create.Error(names.LexV2Models, create.ErrActionCheckingDestroyed, tflexv2models.ResNameBot, rs.Primary.ID, errors.New("not destroyed"))
		}

		return nil
	}
}

func testAccCheckBotExists(ctx context.Context, name string, bot *lexmodelsv2.DescribeBotOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBot, name, errors.New("not found"))
		}

		if rs.Primary.ID == "" {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBot, name, errors.New("not set"))
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)
		resp, err := tflexv2models.FindBotByID(ctx, conn, rs.Primary.ID)
		if err != nil {
			return create.Error(names.LexV2Models, create.ErrActionCheckingExistence, tflexv2models.ResNameBot, rs.Primary.ID, err)
		}

		*bot = *resp

		return nil
	}
}

func testAccPreCheck(ctx context.Context, t *testing.T) {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LexV2ModelsClient(ctx)

	input := &lexmodelsv2.ListBotsInput{}
	_, err := conn.ListBots(ctx, input)

	if acctest.PreCheckSkipError(err) {
		t.Skipf("skipping acceptance testing: %s", err)
	}
	if err != nil {
		t.Fatalf("unexpected PreCheck error: %s", err)
	}
}

// func testAccCheckBotNotRecreated(before, after *lexmodelsv2.DescribeBotOutput) resource.TestCheckFunc {
// 	return func(s *terraform.State) error {
// 		if before, after := aws.ToString(before.BotId), aws.ToString(after.BotId); before != after {
// 			return create.Error(names.LexV2Models, create.ErrActionCheckingNotRecreated, tflexv2models.ResNameBot, aws.ToString(before.BotId), errors.New("recreated"))
// 		}

// 		return nil
// 	}
// }

func testAccBotConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  name                    = %[1]q
  idle_session_ttl_in_seconds = "5"
  role_arn                    = "bot_role_arn"

  data_privacy {
    child_directed = true
  }
}
`, rName)
}

func testAccBotConfig_optional(rName, description, botType, aliasId, aliasName, memberId, memberName string) string {
	return fmt.Sprintf(`
resource "aws_lexv2models_bot" "test" {
  bot_name                    = %[1]q
  description                 = %[2]q
  idle_session_ttl_in_seconds = "5"
  type                        = %[3]q
  role_arn                    = "bot_role_arn"

  data_privacy {
    child_directed = true
  }

  members {
    alias_id   = %[4]q
    alias_name = %[5]q
    id         = %[6]q
    name       = %[7]q
    version    = "2.0"
  }
}
`, rName, description, botType, aliasId, aliasName, memberId, memberName)
}
