package aws

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/tfresource"
)

func TestAccAwsConnectBotAssociationDataSource_basic(t *testing.T) {
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_bot_association.test"
	datasourceName := "data.aws_connect_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccAwsConnectBotAssociationDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectBotAssociationDataSourceConfigBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "bot_name", resourceName, "bot_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "lex_region", resourceName, "lex_region"),
				),
			},
		},
	})
}

func testAccAwsConnectBotAssociationDataSourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_bot_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Connect Bot V1 Association ID not set")
		}
		instanceID, name, _, err := resourceAwsConnectBotV1AssociationParseID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		lexBot, err := finder.BotAssociationV1ByNameWithContext(context.Background(), conn, instanceID, name)

		if errors.Is(err, tfresource.ErrEmptyResult) {
			log.Printf("[DEBUG] Connect Bot V1 Association (%s) not found, has been removed from state", name)
			return nil
		}

		if err != nil {
			return fmt.Errorf("error finding Connect Bot V1 Association by name (%s) potentially still exists: %w", name, err)
		}

		if lexBot != nil {
			return fmt.Errorf("error Connect Bot V1 Association by name (%s): still exists", name)
		}
	}
	return nil
}

func testAccAwsConnectBotAssociationDataSourceBaseConfig(rName string, rName2 string) string {
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
  bot_name    = aws_lex_bot.test.name
  lex_region  = data.aws_region.current.name
}
`, rName, rName2)
}

func testAccAwsConnectBotAssociationDataSourceConfigBasic(rName string, rName2 string) string {
	return fmt.Sprintf(testAccAwsConnectBotAssociationDataSourceBaseConfig(rName, rName2) + `
data "aws_connect_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  bot_name    = aws_connect_bot_association.test.bot_name
}
`)
}
