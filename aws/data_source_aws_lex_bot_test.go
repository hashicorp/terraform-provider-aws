package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceLexBot(t *testing.T) {
	resourceName := "aws_lex_bot.test"
	testId := "test_bot_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexBotConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkLexBotDataSource(resourceName),
				),
			},
		},
	})
}

func checkLexBotDataSource(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualResource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		// Ensure the state is populated with all attributes defined for the resource.
		expectedResource := dataSourceAwsLexBot()
		for k := range expectedResource.Schema {
			if k == "version_or_alias" {
				continue
			}

			if _, ok := actualResource.Primary.Attributes[k]; !ok {
				return fmt.Errorf("state missing attribute %s", k)
			}
		}

		return nil
	}
}

const testDataSourceLexBotConfig = `
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

  name             = "%s"
}

data "aws_lex_bot" "test" {
  name             = "${aws_lex_bot.test.name}"
  version_or_alias = "$LATEST"
}
`
