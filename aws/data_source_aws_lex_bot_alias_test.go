package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceLexBotAlias(t *testing.T) {
	resourceName := "aws_lex_bot_alias.test"
	testId := "test_bot_alias_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexBotAliasConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkLexBotAliasDataSource(resourceName),
				),
			},
		},
	})
}

func checkLexBotAliasDataSource(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualResource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		// Ensure the state is populated with all attributes defined for the resource.
		expectedResource := dataSourceAwsLexBotAlias()
		for k := range expectedResource.Schema {
			if _, ok := actualResource.Primary.Attributes[k]; !ok {
				return fmt.Errorf("state missing attribute %s", k)
			}
		}

		return nil
	}
}

const testDataSourceLexBotAliasConfig = `
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
  name        = "%[1]s"
}

data "aws_lex_bot_alias" "test" {
  bot_name = "${aws_lex_bot.test.name}"
  name     = "${aws_lex_bot_alias.test.name}"
}
`
