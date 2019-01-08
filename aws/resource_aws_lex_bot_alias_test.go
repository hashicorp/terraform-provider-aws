package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLexBotAlias(t *testing.T) {
	resourceName := "aws_lex_bot_alias.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testBotId := "test_bot_" + testId
	testBotAliasId := "test_bot_alias_" + testId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexBotAliasDestroy(testBotAliasId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexBotAliasConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet(resourceName, "bot_name"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_version"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexBotAlias()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%s.%s", testBotId, testBotAliasId),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testLexBotAliasUpdateConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias update."),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexBotAlias()),
				),
			},
		},
	})
}

func checkLexBotAliasDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(id),
			VersionOrAlias: aws.String("$LATEST"),
		})

		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return nil
			}

			return fmt.Errorf("could not get Lex bot alias, %s", id)
		}

		return fmt.Errorf("bot alias still exists after delete, %s", id)
	}
}

const testLexBotAliasConfig = `
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

const testLexBotAliasUpdateConfig = `
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
