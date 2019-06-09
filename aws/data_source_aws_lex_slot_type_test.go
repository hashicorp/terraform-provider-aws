package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
)

func TestAccDataSourceAwsLexSlotType(t *testing.T) {
	resourceName := "aws_lex_slot_type.test"
	dataSourceName := "data." + resourceName
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceAwsLexSlotTypeConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "checksum", resourceName, "checksum"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "value_selection_strategy", resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrPair(dataSourceName, "version", resourceName, "version"),
					resource.TestCheckResourceAttrSet(dataSourceName, "created_date"),
					resource.TestCheckResourceAttrSet(dataSourceName, "last_updated_date"),
				),
			},
		},
	})
}

const testDataSourceAwsLexSlotTypeConfig = `
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"
  name        = "test_slot_type_%s"
}

data "aws_lex_slot_type" "test" {
  name    = "${aws_lex_slot_type.test.name}"
  version = "${aws_lex_slot_type.test.version}"
}
`
