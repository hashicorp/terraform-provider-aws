package aws

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/connect"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/terraform-providers/terraform-provider-aws/aws/internal/service/connect/finder"
)

//Serialized acceptance tests due to Connect account limits (max 2 parallel tests)
func TestAccAwsConnectLexBotAssociation_serial(t *testing.T) {
	testCases := map[string]func(t *testing.T){
		"basic":      testAccAwsConnectLexBotAssociation_basic,
		"disappears": testAccAwsConnectLexBotAssociation_disappears,
	}

	for name, tc := range testCases {
		tc := tc
		t.Run(name, func(t *testing.T) {
			tc(t)
		})
	}
}

func testAccAwsConnectLexBotAssociation_basic(t *testing.T) {
	var v connect.LexBot
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource_test_terraform")
	resourceName := "aws_connect_lex_bot_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectLexBotAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLexBotAssociationConfigBasic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectLexBotAssociationExists(resourceName, &v),
					resource.TestCheckResourceAttrSet(resourceName, "instance_id"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "region"),
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

func testAccAwsConnectLexBotAssociation_disappears(t *testing.T) {
	var v connect.LexBot
	rName := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	rName2 := acctest.RandomWithPrefix("resource_test_terraform")
	resourceName := "aws_connect_lex_bot_association.test"
	instanceResourceName := "aws_connect_lex_bot_association.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		ErrorCheck:   testAccErrorCheck(t, connect.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsConnectLexBotAssociationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsConnectLexBotAssociationConfigBasic(rName, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsConnectLexBotAssociationExists(resourceName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsConnectLexBotAssociation(), instanceResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsConnectLexBotAssociationExists(resourceName string, function *connect.LexBot) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("Connect Lex Bot Association not found: %s", resourceName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Lex Bot Association ID not set")
		}
		instanceID, name, _, err := resourceAwsConnectLexBotAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		lexBot, err := finder.LexBotAssociationByName(context.Background(), conn, instanceID, name)
		if err != nil {
			return fmt.Errorf("error finding LexBot Association by name (%s): %w", name, err)
		}

		if lexBot == nil {
			return fmt.Errorf("error finding LexBot Association by name (%s): not found", name)
		}

		return nil
	}
}

func testAccCheckAwsConnectLexBotAssociationDestroy(s *terraform.State) error {
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_connect_lex_bot_association" {
			continue
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("Connect Lex Bot Association ID not set")
		}
		instanceID, name, _, err := resourceAwsConnectLexBotAssociationParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := testAccProvider.Meta().(*AWSClient).connectconn

		lexBot, err := finder.LexBotAssociationByName(context.Background(), conn, instanceID, name)
		if err == nil {
			return fmt.Errorf("error finding LexBot Association by name (%s): still exists", name)
		}

		if lexBot != nil {
			return fmt.Errorf("error LexBot Association by name (%s): still exists", name)
		}
	}
	return nil
}

func testAccAwsConnectLexBotAssociationConfigBase(rName string, rName2 string) string {
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

func testAccAwsConnectLexBotAssociationConfigBasic(rName string, rName2 string) string {
	return composeConfig(
		testAccAwsConnectLexBotAssociationConfigBase(rName, rName2),
		fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_connect_lex_bot_association" "test" {
  instance_id = aws_connect_instance.test.id
  bot_name    = %[1]q
  lex_region  = "${data.aws_region.current.name}"
}
`, rName))
}
