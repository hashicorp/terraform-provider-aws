// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBotAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAssociationConfig_v1Basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttr(resourceName, "lex_bot.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "lex_bot.0.name", "aws_lex_bot.test", names.AttrName),
					resource.TestCheckResourceAttrSet(resourceName, "lex_bot.0.lex_region"),
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

func testAccBotAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_bot_association.test"
	instanceResourceName := "aws_connect_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckBotAssociationDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccBotAssociationConfig_v1Basic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckBotAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfconnect.ResourceBotAssociation(), instanceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckBotAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Bot Association not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Bot Association ID not set")
		}
		instanceID, name, region, err := tfconnect.BotV1AssociationParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

		lexBot, err := tfconnect.FindBotAssociationV1ByNameAndRegionWithContext(ctx, conn, instanceID, name, region)

		if err != nil {
			return fmt.Errorf("error finding Connect Bot Association (%s): %w", rs.Primary.ID, err)
		}

		if lexBot == nil {
			return fmt.Errorf("error finding Connect Bot Association (%s): not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckBotAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_connect_bot_association" {
				continue
			}

			if rs.Primary.ID == "" {
				return fmt.Errorf("Connect Connect Bot V1 Association ID not set")
			}

			instanceID, name, region, err := tfconnect.BotV1AssociationParseResourceID(rs.Primary.ID)
			if err != nil {
				return err
			}

			conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn(ctx)

			lexBot, err := tfconnect.FindBotAssociationV1ByNameAndRegionWithContext(ctx, conn, instanceID, name, region)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return fmt.Errorf("error finding Connect Bot Association (%s): %w", rs.Primary.ID, err)
			}

			if lexBot != nil {
				return fmt.Errorf("Connect Bot Association (%s) still exists", rs.Primary.ID)
			}
		}
		return nil
	}
}

func testAccBotV1AssociationConfigBase(rName, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_lex_intent" "test" {
  create_version = true
  name           = %[1]q
  fulfillment_activity {
    type = "ReturnIntent"
  }
  sample_utterances = [
    "I would like to pick up flowers",
  ]
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
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
    intent_version = "1"
  }
  child_directed   = false
  name             = %[1]q
  process_behavior = "BUILD"
}

resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[2]q
  outbound_calls_enabled   = true
}
  `, rName, rName2)
}

func testAccBotAssociationConfig_v1Basic(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccBotV1AssociationConfigBase(rName, rName2),
		`
resource "aws_connect_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  lex_bot {
    name = aws_lex_bot.test.name
  }
}
`)
}
