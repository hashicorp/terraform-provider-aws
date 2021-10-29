package connect_test

import (
	"context"
	"errors"
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfconnect "github.com/hashicorp/terraform-provider-aws/internal/service/connect"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccBotAssociationDataSource_basic(t *testing.T) {
	rName := sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	rName2 := sdkacctest.RandomWithPrefix("resource-test-terraform")
	resourceName := "aws_connect_bot_association.test"
	datasourceName := "data.aws_connect_bot_association.test"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, connect.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccBotAssociationDataSourceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccBotAssociationDataSourceConfigBasic(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "instance_id", resourceName, "instance_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "bot_name", resourceName, "bot_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "lex_region", resourceName, "lex_region"),
				),
			},
		},
	})
}

func testAccBotAssociationDataSourceDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_bot_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Connect Bot V1 Association ID not set")
		}
		instanceID, name, _, err := tfconnect.BotV1AssociationParseResourceID(rs.Primary.ID)
		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).ConnectConn

		lexBot, err := tfconnect.FindBotAssociationV1ByNameWithContext(context.Background(), conn, instanceID, name)

		if tfawserr.ErrCodeEquals(err, tfconnect.BotAssociationStatusNotFound, "") || errors.Is(err, tfresource.ErrEmptyResult) {
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

func testAccBotAssociationDataSourceBaseConfig(rName string, rName2 string) string {
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

func testAccBotAssociationDataSourceConfigBasic(rName string, rName2 string) string {
	return fmt.Sprintf(testAccBotAssociationDataSourceBaseConfig(rName, rName2) + `
data "aws_connect_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  bot_name    = aws_connect_bot_association.test.bot_name
}
`)
}
