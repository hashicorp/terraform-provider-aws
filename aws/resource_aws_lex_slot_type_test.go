package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLexSlotType(t *testing.T) {
	resourceName := "aws_lex_slot_type.test"
	testID := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testSlotTypeID := "test_slot_type_" + testID

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeDestroy(testSlotTypeID, "$LATEST"),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testAccAwsLexSlotTypeConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, "$LATEST"),

					// user defined attributes
					resource.TestCheckResourceAttr(resourceName, "description", "Types of flowers to pick up"),
					resource.TestCheckResourceAttr(resourceName, "enumeration_value.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "name", testSlotTypeID),
					resource.TestCheckResourceAttr(resourceName, "value_selection_strategy", "ORIGINAL_VALUE"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testAccAwsLexSlotTypeUpdateConfig, testID),
				Check: resource.ComposeAggregateTestCheckFunc(
					// user defined attributes
					resource.TestCheckResourceAttr(resourceName, "description", "Allowed flower types"),
					resource.TestCheckResourceAttr(resourceName, "enumeration_value.#", "2"),
					resource.TestCheckResourceAttr(resourceName, "name", testSlotTypeID),
					resource.TestCheckResourceAttr(resourceName, "value_selection_strategy", "TOP_RESOLUTION"),

					// computed attributes
					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
				),
			},
		},
	})
}

func testAccCheckAwsLexSlotTypeExists(slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})
		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return fmt.Errorf("error slot type %s not found, %s", slotTypeName, err)
			}

			return fmt.Errorf("error getting slot type %s: %s", slotTypeName, err)
		}

		return nil
	}
}

func testAccCheckAwsLexSlotTypeDestroy(slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})

		if err != nil {
			if isAWSErr(err, "NotFoundException", "") {
				return nil
			}

			return fmt.Errorf("error getting slot type %s: %s", slotTypeName, err)
		}

		return fmt.Errorf("error slot type still exists after delete, %s", slotTypeName)
	}
}

const testAccAwsLexSlotTypeConfig = `
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"

  enumeration_value {
    synonyms = [
      "Lirium",
    ]

    value = "lilies"
  }

  name                     = "test_slot_type_%s"
  value_selection_strategy = "ORIGINAL_VALUE"
}
`

const testAccAwsLexSlotTypeUpdateConfig = `
resource "aws_lex_slot_type" "test" {
  description = "Allowed flower types"

  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]

    value = "lilies"
  }

  enumeration_value {
    synonyms = [
      "Eduardoregelia",
      "Podonix",
    ]

    value = "tulips"
  }

  name                     = "test_slot_type_%s"
  value_selection_strategy = "TOP_RESOLUTION"
}
`
