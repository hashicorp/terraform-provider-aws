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

func TestAccAwsLexBot(t *testing.T) {
	resourceName := "aws_lex_bot.test"
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testBotID := "test_bot_" + testID

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexBotDestroy(testBotID, "$LATEST"),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsLexBotConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexBotExists(testBotID, "$LATEST"),

					// user provided attributes
					resource.TestCheckResourceAttr(resourceName, "abort_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "child_directed", "false"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.max_attempts", "2"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "description", "Bot to order flowers on the behalf of a user"),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "600"),
					resource.TestCheckResourceAttr(resourceName, "intent.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-US"),
					resource.TestCheckResourceAttr(resourceName, "name", testBotID),
					resource.TestCheckResourceAttr(resourceName, "process_behavior", "SAVE"),
					resource.TestCheckResourceAttr(resourceName, "voice_id", "Salli"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
			{
				Config: fmt.Sprintf(testAccAwsLexBotUpdateConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(

					// user provided attributes
					resource.TestCheckResourceAttr(resourceName, "abort_statement.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "abort_statement.0.message.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "child_directed", "true"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.max_attempts", "3"),
					resource.TestCheckResourceAttr(resourceName, "clarification_prompt.0.message.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "description", "Bot to order flowers"),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "intent.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "locale", "en-US"),
					resource.TestCheckResourceAttr(resourceName, "name", testBotID),
					resource.TestCheckResourceAttr(resourceName, "process_behavior", "BUILD"),
					resource.TestCheckResourceAttr(resourceName, "voice_id", "Ivy"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
				),
			},
		},
	})
}

func testAccCheckAwsLexBotExists(botName, botVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(botName),
			VersionOrAlias: aws.String(botVersion),
		})
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return fmt.Errorf("error bot %s not found, %s", botName, err)
			}

			return fmt.Errorf("error getting bot %s: %s", botName, err)
		}

		return nil
	}
}

func testAccCheckAwsLexBotDestroy(botName, botVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetBot(&lexmodelbuildingservice.GetBotInput{
			Name:           aws.String(botName),
			VersionOrAlias: aws.String(botVersion),
		})

		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}

			return fmt.Errorf("error getting bot %s: %s", botName, err)
		}

		return fmt.Errorf("error bot still exists after delete, %s", botName)
	}
}

const testAccAwsLexBotConfig = `
resource "aws_lex_intent" "test_intent_one" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_one_%[1]s"
}

resource "aws_lex_intent" "test_intent_two" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_two_%[1]s"
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
`

const testAccAwsLexBotUpdateConfig = `
resource "aws_lex_intent" "test_intent_one" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_one_%[1]s"
}

resource "aws_lex_intent" "test_intent_two" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_two_%[1]s"
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I am not able to assist at this time"
      content_type = "PlainText"
    }

    message {
      content      = "Sorry, I can't complete your request"
      content_type = "PlainText"
    }
  }

  child_directed = true

  clarification_prompt {
    max_attempts = 3

    message {
      content      = "I didn't understand you, what would you like to do?"
      content_type = "PlainText"
    }

    message {
      content      = "I'm sorry, I don't understand your request, what would you like to do?"
      content_type = "PlainText"
    }
  }

  description = "Bot to order flowers"

  idle_session_ttl_in_seconds = 300

  intent {
    intent_name    = "${aws_lex_intent.test_intent_one.name}"
    intent_version = "${aws_lex_intent.test_intent_one.version}"
  }

  intent {
    intent_name    = "${aws_lex_intent.test_intent_two.name}"
    intent_version = "${aws_lex_intent.test_intent_two.version}"
  }

  locale           = "en-US"
  name             = "test_bot_%[1]s"
  process_behavior = "BUILD"
  voice_id         = "Ivy"
}
`
