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

func TestAccLexBot(t *testing.T) {
	resourceName := "aws_lex_bot.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testBotId := "test_bot_" + testId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexBotDestroy(testBotId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexBotConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrPrefixSet(resourceName, "abort_statement"),
					testCheckResourceAttrPrefixSet(resourceName, "clarification_prompt"),
					testCheckResourceAttrPrefixSet(resourceName, "intent"),

					resource.TestCheckResourceAttrSet(resourceName, "child_directed"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "idle_session_ttl_in_seconds"),
					resource.TestCheckResourceAttrSet(resourceName, "locale"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "process_behavior"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "voice_id"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexBot()),
				),
			},
			{
				ResourceName: resourceName,
				ImportState:  true,
			},
			{
				Config: fmt.Sprintf(testLexBotUpdateConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "child_directed", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Bot to order flowers"),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "process_behavior", "BUILD"),
					resource.TestCheckResourceAttr(resourceName, "voice_id", "Ivy"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexBot()),
				),
			},
		},
	})
}

func checkLexBotDestroy(id string) resource.TestCheckFunc {
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

			return fmt.Errorf("could not get Lex bot, %s", id)
		}

		return fmt.Errorf("bot still exists after delete, %s", id)
	}
}

const testLexBotConfig = `
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

const testLexBotUpdateConfig = `
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
