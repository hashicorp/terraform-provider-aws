package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsLexBotAlias(t *testing.T) {
	resourceName := "aws_lex_bot_alias.test"
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testBotID := "test_bot_" + testID
	testBotAliasID := "test_bot_alias_" + testID

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexBotAliasDestroy(testBotID, testBotAliasID),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsLexBotAliasConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexBotAliasExists(testBotID, testBotAliasID),

					// user provided attributes
					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias create."),
					resource.TestCheckResourceAttr(resourceName, "bot_name", testBotID),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "name", testBotAliasID),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s.%s", testBotID, testBotAliasID),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testAccAwsLexBotAliasUpdateConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// user updated attributes
					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias update."),
					resource.TestCheckResourceAttr(resourceName, "bot_name", testBotID),
					resource.TestCheckResourceAttr(resourceName, "bot_version", "$LATEST"),
					resource.TestCheckResourceAttr(resourceName, "name", testBotAliasID),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
				),
			},
		},
	})
}

func testAccCheckAwsLexBotAliasExists(botName, botAliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})
		if err != nil {
			return fmt.Errorf("error getting bot alias %s: %s", botAliasName, err)
		}

		return nil
	}
}

func testAccCheckAwsLexBotAliasDestroy(botName, botAliasName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBotAlias(&lexmodelbuildingservice.GetBotAliasInput{
			BotName: aws.String(botName),
			Name:    aws.String(botAliasName),
		})

		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}

			return fmt.Errorf("error getting bot alias %s: %s", botAliasName, err)
		}

		return fmt.Errorf("error bot alias still exists after delete, %s", botAliasName)
	}
}

const testAccAwsLexBotAliasConfig = `
resource "aws_lex_intent" "test_intent_one" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_one_%[1]s"
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }
  }

  child_directed = false

  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }

  description = "Bot to order flowers on the behalf of a user"

  idle_session_ttl_in_seconds = 600

  intent {
    intent_name    = "${aws_lex_intent.test_intent_one.name}"
    intent_version = "${aws_lex_intent.test_intent_one.version}"
  }

  locale           = "en-US"
  name             = "test_bot_%[1]s"
  process_behavior = "SAVE"
  voice_id         = "Salli"
}

resource "aws_lex_bot_alias" "test" {
  bot_name    = "${aws_lex_bot.test.name}"
  bot_version = "${aws_lex_bot.test.version}"
  description = "Testing lex bot alias create."
  name        = "test_bot_alias_%[1]s"
}
`

const testAccAwsLexBotAliasUpdateConfig = `
resource "aws_lex_intent" "test_intent_one" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_one_%[1]s"
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }
  }

  child_directed = false

  clarification_prompt {
    max_attempts = 2

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }
  }

  description = "Bot to order flowers on the behalf of a user"

  idle_session_ttl_in_seconds = 600

  intent {
    intent_name    = "${aws_lex_intent.test_intent_one.name}"
    intent_version = "${aws_lex_intent.test_intent_one.version}"
  }

  locale           = "en-US"
  name             = "test_bot_%[1]s"
  process_behavior = "SAVE"
  voice_id         = "Salli"
}

resource "aws_lex_bot_alias" "test" {
  bot_name    = "${aws_lex_bot.test.name}"
  bot_version = "${aws_lex_bot.test.version}"
  description = "Testing lex bot alias update."
  name        = "test_bot_alias_%[1]s"
}
`
