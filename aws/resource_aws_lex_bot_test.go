package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func testCheckResourceAttrPrefixSet(resourceName, prefix string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rm := s.RootModule()
		rs, ok := rm.Resources[resourceName]

		if !ok {
			return fmt.Errorf("resource does not exist in state, %s", resourceName)
		}

		for attr := range rs.Primary.Attributes {
			if strings.HasPrefix(attr, prefix+".") {
				return nil
			}
		}

		return fmt.Errorf("resource attribute prefix does not exist in state, %s", prefix)
	}
}

func TestAccLexBot(t *testing.T) {
	resourceName := "aws_lex_bot.test"
	testId := "test_bot_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexBotDestroy(testId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexBotConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					// Validate AWS state

					checkLexBotCreate(testId),

					// Validate Terraform state

					testCheckResourceAttrPrefixSet(resourceName, "abort_statement"),
					testCheckResourceAttrPrefixSet(resourceName, "clarification_prompt"),
					testCheckResourceAttrPrefixSet(resourceName, "intent"),

					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "child_directed"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "idle_session_ttl_in_seconds"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrSet(resourceName, "locale"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "process_behavior"),
					resource.TestCheckResourceAttrSet(resourceName, "status"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
					resource.TestCheckResourceAttrSet(resourceName, "voice_id"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testLexBotUpdateConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					// Validate AWS state

					checkLexBotUpdate(testId),

					// Validate Terraform state

					resource.TestCheckResourceAttr(resourceName, "child_directed", "true"),
					resource.TestCheckResourceAttr(resourceName, "description", "Bot to order flowers"),
					resource.TestCheckResourceAttr(resourceName, "idle_session_ttl_in_seconds", "300"),
					resource.TestCheckResourceAttr(resourceName, "process_behavior", "BUILD"),
					resource.TestCheckResourceAttr(resourceName, "voice_id", "Ivy"),
				),
			},
		},
	})
}

func getLexBot(id string) (*lexmodelbuildingservice.GetBotOutput, error) {
	conn := testAccProvider.Meta().(*AWSClient).lexmodelconn
	return conn.GetBot(&lexmodelbuildingservice.GetBotInput{
		Name:           aws.String(id),
		VersionOrAlias: aws.String("$LATEST"),
	})
}

func checkLexBotCreate(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bot, err := getLexBot(id)
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return fmt.Errorf("bot does not exist, %s", id)
			}

			return fmt.Errorf("could not get Lex bot, %s", id)
		}

		if bot.AbortStatement == nil {
			return fmt.Errorf("bot has no abort statement")
		}
		if len(bot.AbortStatement.Messages) == 0 {
			return fmt.Errorf("bot abort statement has no messages")
		}

		if bot.ClarificationPrompt == nil {
			return fmt.Errorf("bot has no clarification prompt")
		}
		if len(bot.ClarificationPrompt.Messages) == 0 {
			return fmt.Errorf("bot clarification prompt has no messages")
		}

		if bot.Intents == nil || len(bot.Intents) == 0 {
			return fmt.Errorf("bot has no intents")
		}

		return nil
	}
}

func checkLexBotUpdate(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		bot, err := getLexBot(id)
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return fmt.Errorf("bot does not exist, %s", id)
			}

			return fmt.Errorf("could not get Lex bot, %s", id)
		}

		if bot.AbortStatement == nil {
			return fmt.Errorf("bot has no abort statement")
		}
		if len(bot.AbortStatement.Messages) != 2 {
			return fmt.Errorf(
				"bot abort statement has %d messages, expected 2", len(bot.AbortStatement.Messages))
		}

		if bot.ClarificationPrompt == nil {
			return fmt.Errorf("bot has no clarification prompt")
		}
		if len(bot.ClarificationPrompt.Messages) != 2 {
			return fmt.Errorf(
				"bot clarification prompt has %d messages, expected 2", len(bot.ClarificationPrompt.Messages))
		}

		if bot.Intents == nil || len(bot.Intents) != 2 {
			return fmt.Errorf("bot has %d intents, expected 2", len(bot.Intents))
		}

		return nil
	}
}

func checkLexBotDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLexBot(id)
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
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  locale           = "en-US"
  name             = "%s"
  process_behavior = "SAVE"
  voice_id         = "Salli"
}
`

const testLexBotUpdateConfig = `
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
    intent_name    = "OrderFlowers"
    intent_version = "2"
  }
  intent {
    intent_name    = "GetOrderStatus"
    intent_version = "1"
  }

  locale           = "en-US"
  name             = "%s"
  process_behavior = "BUILD"
  voice_id         = "Ivy"
}
`
