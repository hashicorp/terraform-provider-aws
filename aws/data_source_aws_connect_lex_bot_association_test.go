package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAwsConnectLexBotAssociationDataSource_Name(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_lex_bot_association.test"
	datasourceName := "data.aws_connect_lex_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:   func() { testAccPreCheck(t) },
		ErrorCheck: testAccErrorCheck(t, connect.EndpointsID),
		Providers:  testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLexBotAssociationDataSourceConfig_Name(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(datasourceName, "region", resourceName, "region"),
				),
			},
		},
	})
}

func testAccAwsConnectLexBotAssociationDataSourceBaseConfig(rName string, rName2 string) string {
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

resource "aws_connect_lex_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  name        = %[1]q
  region      = "${data.aws_region.current.name}"
}
`, rName, rName2)
}

func testAccAwsConnectLexBotAssociationDataSourceConfig_Name(rName string, rName2 string) string {
	return fmt.Sprintf(testAccAwsConnectLexBotAssociationDataSourceBaseConfig(rName, rName2) + `
data "aws_connect_lex_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  name        = aws_connect_lex_bot_association.test.name
}
`)
}
