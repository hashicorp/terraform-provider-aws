package aws

import (
	"fmt"
	"strings"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/helper/schema"
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

func checkResourceStateComputedAttr(resourceName string, expectedResource *schema.Resource) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualResource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		// Ensure the state is populated with all the computed attributes defined by the resource schema.
		for k, v := range expectedResource.Schema {
			if !v.Computed {
				continue
			}

			if _, ok := actualResource.Primary.Attributes[k]; !ok {
				return fmt.Errorf("state missing attribute %s", k)
			}
		}

		return nil
	}
}

func TestAccDataSourceLexBot(t *testing.T) {
	resourceName := "aws_lex_bot.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexBotConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkResourceStateComputedAttr(resourceName, dataSourceAwsLexBot()),
				),
			},
		},
	})
}

const testDataSourceLexBotConfig = `
resource "aws_lex_intent" "test" {
  fulfillment_activity {
    type = "ReturnIntent"
  }

  name = "test_intent_%[1]s"
}

resource "aws_lex_bot" "test" {
  abort_statement {
    message {
      content      = "Sorry, I'm not able to assist at this time"
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

  intent {
    intent_name    = "${aws_lex_intent.test.name}"
    intent_version = "${aws_lex_intent.test.version}"
  }

  name     = "test_bot_%[1]s"
  voice_id = "Salli"
}

data "aws_lex_bot" "test" {
  name    = "${aws_lex_bot.test.name}"
  version = "${aws_lex_bot.test.version}"
}
`
