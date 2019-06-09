package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsLexBotAlias(t *testing.T) {
	resourceName := "aws_lex_bot_alias.test"
	dataSourceName := "data." + resourceName
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceAwsLexBotAliasConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_name", resourceName, "bot_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bot_version", resourceName, "bot_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_date"),
				),
			},
		},
	})
}

const testDataSourceAwsLexBotAliasConfig = `
resource "aws_lex_intent" "test" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"
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

  name = "test_bot_%[1]s"

  intent {
    intent_name    = "${aws_lex_intent.test.name}"
    intent_version = "${aws_lex_intent.test.version}"
  }
}

resource "aws_lex_bot_alias" "test" {
  bot_name    = "${aws_lex_bot.test.name}"
  bot_version = "${aws_lex_bot.test.version}"
  name        = "test_bot_alias_%[1]s"
  description = "A test bot alias"
}

data "aws_lex_bot_alias" "test" {
  bot_name = "${aws_lex_bot.test.name}"
  name     = "${aws_lex_bot_alias.test.name}"
}
`
