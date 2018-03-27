package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccDataSourceLexSlotType(t *testing.T) {
	resourceName := "aws_lex_slot_type.test"
	testId := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testDataSourceLexSlotTypeConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					checkLexSlotTypeDataSource(resourceName),
				),
			},
		},
	})
}

func checkLexSlotTypeDataSource(resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		actualResource, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("root module has no resource called %s", resourceName)
		}

		// Ensure the state is populated with all attributes defined for the resource.
		expectedResource := dataSourceAwsLexSlotType()
		for k := range expectedResource.Schema {
			if _, ok := actualResource.Primary.Attributes[k]; !ok {
				return fmt.Errorf("state missing attribute %s", k)
			}
		}

		return nil
	}
}

const testDataSourceLexSlotTypeConfig = `
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"

  enumeration_value {
    synonyms = [
      "Lirium",
    ]
    value    = "lilies"
  }

  name                     = "%s"
  value_selection_strategy = "ORIGINAL_VALUE"
}

data "aws_lex_slot_type" "test" {
  name    = "${aws_lex_slot_type.test.name}"
  version = "$LATEST"
}
`
