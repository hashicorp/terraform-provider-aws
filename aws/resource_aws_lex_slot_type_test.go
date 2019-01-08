package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccLexSlotType(t *testing.T) {
	resourceName := "aws_lex_slot_type.test"
	testId := acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testSlotTypeId := "test_slot_type_" + testId

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexSlotTypeDestroy(testSlotTypeId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexSlotTypeConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					testCheckResourceAttrPrefixSet(resourceName, "enumeration_value"),

					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexSlotType()),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: fmt.Sprintf(testLexSlotTypeUpdateConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "description", "Allowed flower types"),
					resource.TestCheckResourceAttr(resourceName, "value_selection_strategy", "TOP_RESOLUTION"),

					checkResourceStateComputedAttr(resourceName, resourceAwsLexSlotType()),
				),
			},
		},
	})
}

func checkLexSlotTypeDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(id),
			Version: aws.String("$LATEST"),
		})

		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return nil
			}

			return fmt.Errorf("could not get Lex slot type, %s", id)
		}

		return fmt.Errorf("slot type still exists after delete, %s", id)
	}
}

const testLexSlotTypeConfig = `
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

const testLexSlotTypeUpdateConfig = `
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
