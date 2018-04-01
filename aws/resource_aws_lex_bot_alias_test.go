package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLexBotAlias(t *testing.T) {
	resourceName := "aws_lex_bot_alias.test"
	testId := "test_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexBotAliasDestroy(testId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexBotAliasConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					// Validate Terraform state

					resource.TestCheckResourceAttrSet(resourceName, "bot_name"),
					resource.TestCheckResourceAttrSet(resourceName, "bot_version"),
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportStateId:     fmt.Sprintf("%[1]s.%[1]s", testId),
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testLexBotAliasUpdateConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					// Validate Terraform state

					resource.TestCheckResourceAttr(resourceName, "description", "Testing lex bot alias update."),
				),
			},
		},
	})
}

func checkLexBotAliasDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLexBot(id)
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

  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  name             = "%[1]s"
}

resource "aws_lex_bot_alias" "test" {
  bot_name    = "${aws_lex_bot.test.name}"
  bot_version = "${aws_lex_bot.test.version}"
  description = "Testing lex bot alias create."
  name        = "%[1]s"

  depends_on  = ["aws_lex_bot.test"]
}
`

const testLexBotAliasUpdateConfig = `
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

  intent {
    intent_name    = "OrderFlowers"
    intent_version = "1"
  }

  name             = "%[1]s"
  process_behavior = "SAVE"
}

resource "aws_lex_bot_alias" "test" {
  bot_name    = "${aws_lex_bot.test.name}"
  bot_version = "${aws_lex_bot.test.version}"
  description = "Testing lex bot alias update."
  name        = "%[1]s"

  depends_on  = ["aws_lex_bot.test"]
}
`
