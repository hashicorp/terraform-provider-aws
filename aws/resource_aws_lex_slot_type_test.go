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
	testId := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: checkLexSlotTypeDestroy(testId),
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(testLexSlotTypeConfig, testId),
				Check: resource.ComposeTestCheckFunc(
					// Validate AWS state

					checkLexSlotTypeCreate(testId),

					// Validate Terraform state

					testCheckResourceAttrPrefixSet(resourceName, "enumeration_value"),

					resource.TestCheckResourceAttrSet(resourceName, "checksum"),
					resource.TestCheckResourceAttrSet(resourceName, "created_date"),
					resource.TestCheckResourceAttrSet(resourceName, "description"),
					resource.TestCheckResourceAttrSet(resourceName, "last_updated_date"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttrSet(resourceName, "value_selection_strategy"),
					resource.TestCheckResourceAttrSet(resourceName, "version"),
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
					// Validate AWS state

					checkLexSlotTypeUpdate(testId),

					// Validate Terraform state

					resource.TestCheckResourceAttr(resourceName, "description", "Allowed flower types"),
					resource.TestCheckResourceAttr(resourceName, "value_selection_strategy", "TOP_RESOLUTION"),
				),
			},
		},
	})
}

func getLexSlotType(id string) (*lexmodelbuildingservice.GetSlotTypeOutput, error) {
	conn := testAccProvider.Meta().(*AWSClient).lexmodelconn
	return conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
		Name:    aws.String(id),
		Version: aws.String("$LATEST"),
	})
}

func checkLexSlotTypeCreate(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		slotType, err := getLexSlotType(id)
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return fmt.Errorf("slot type does not exist, %s", id)
			}

			return fmt.Errorf("could not get Lex slot type, %s", id)
		}

		if slotType.EnumerationValues == nil {
			return fmt.Errorf("slot type EnumerationValues is nil")
		}
		if len(slotType.EnumerationValues) != 1 {
			return fmt.Errorf("slot type should have 1 enumeration value")
		}
		if slotType.EnumerationValues[0].Synonyms == nil {
			return fmt.Errorf("slot type enumeration value Synonyms is nil")
		}
		if len(slotType.EnumerationValues[0].Synonyms) != 1 {
			return fmt.Errorf("slot type enumeration value should have 1 synonym")
		}

		return nil
	}
}

func checkLexSlotTypeUpdate(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		slotType, err := getLexSlotType(id)
		if err != nil {
			aerr, ok := err.(awserr.Error)
			if ok && aerr.Code() == "NotFoundException" {
				return fmt.Errorf("slot type does not exist, %s", id)
			}

			return fmt.Errorf("could not get Lex slot type, %s", id)
		}

		if slotType.EnumerationValues == nil {
			return fmt.Errorf("slot type EnumerationValues is nil")
		}
		if len(slotType.EnumerationValues) != 2 {
			return fmt.Errorf("slot type should have 2 enumeration values")
		}
		if slotType.EnumerationValues[0].Synonyms == nil {
			return fmt.Errorf("slot type enumeration value Synonyms is nil")
		}
		if len(slotType.EnumerationValues[0].Synonyms) != 2 {
			return fmt.Errorf("slot type enumeration value should have 2 synonyms")
		}

		return nil
	}
}

func checkLexSlotTypeDestroy(id string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, err := getLexSlotType(id)
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
    value    = "lilies"
  }

  name                     = "%s"
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

  name                     = "%s"
  value_selection_strategy = "TOP_RESOLUTION"
}
`
