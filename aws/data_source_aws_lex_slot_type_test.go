package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceLexSlotType(t *testing.T) {
	resourceName := "aws_lex_slot_type.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexSlotTypeConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkResourceStateComputedAttr(resourceName, dataSourceAwsLexSlotType()),
				),
			},
		},
	})
}

const testDataSourceLexSlotTypeConfig = `
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"
  name        = "test_slot_type_%s"
}

data "aws_lex_slot_type" "test" {
  name    = "${aws_lex_slot_type.test.name}"
  version = "${aws_lex_slot_type.test.version}"
}
`
