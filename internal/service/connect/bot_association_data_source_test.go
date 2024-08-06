// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccBotAssociationDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_bot_association.test"
	datasourceName := "data.aws_connect_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccBotAssociationDataSourceConfig_basic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, "lex_bot", resourceName, "lex_bot"),
				),
			},
		},
	})
}

func testAccBotAssociationDataSourceConfig_base(rName string, rName2 string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

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

resource "aws_connect_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  lex_bot {
    lex_region = data.aws_region.current.name
    name       = aws_lex_bot.test.name
  }
}
`, rName, rName2)
}

func testAccBotAssociationDataSourceConfig_basic(rName string, rName2 string) string {
	return fmt.Sprintf(testAccBotAssociationDataSourceConfig_base(rName, rName2) + `
data "aws_connect_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  lex_bot {
    name = aws_connect_bot_association.test.lex_bot[0].name
  }
}
`)
}
