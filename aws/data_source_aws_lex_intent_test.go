package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceLexIntent(t *testing.T) {
	resourceName := "aws_lex_intent.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexIntentConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkResourceStateComputedAttr(resourceName, dataSourceAwsLexIntent()),
				),
			},
		},
	})
}

const testDataSourceLexIntentConfig = `
resource "aws_lex_intent" "test" {
  description = "Intent to order a bouquet of flowers for pick up"

  fulfillment_activity {
    type = "ReturnIntent"
  }

  name                    = "test_intent_%s"
  parent_intent_signature = "AMAZON.FallbackIntent"
}

data "aws_lex_intent" "test" {
  name = "${aws_lex_intent.test.name}"
}
`
