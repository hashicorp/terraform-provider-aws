package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

func TestAccLexSlotType_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "description", ""),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "0"),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttrSet(rName, "version"),
					testAccCheckResourceAttrRfc3339(rName, "created_date"),
					testAccCheckResourceAttrRfc3339(rName, "last_updated_date"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_createVersion(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_withoutVersion(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					testAccCheckAwsLexSlotTypeExistsWithVersion(testSlotTypeID, "1", &v),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_description(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "description", ""),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeUpdateConfig_description(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "description", "Types of flowers to pick up"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_enumerationValues(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "1"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_enumerationValues(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "2"),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_name(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID1 := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)
	testSlotTypeID2 := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID1, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID1, &v),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID1),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID2, &v),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID2),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_valueSelectionStrategy(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(testSlotTypeID, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccAwsLexSlotTypeConfig_valueSelectionStrategy(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(testSlotTypeID, &v),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyTopResolution),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
		},
	})
}

func TestAccAwsLexSlotType_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	resourceName := "aws_lex_slot_type.test"
	rName := "test_slot_type_" + acctest.RandStringFromCharSet(8, acctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsLexSlotTypeNotExists(rName, LexSlotTypeVersionLatest),
		Steps: []resource.TestStep{
			{
				Config: testAccAwsLexSlotTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsLexSlotTypeExists(rName, &v),
					testAccCheckResourceDisappears(testAccProvider, resourceAwsLexSlotType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckAwsLexSlotTypeExistsWithVersion(slotTypeName, slotTypeVersion string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		var err error
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		output, err = conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})
		if isAWSErr(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
			return fmt.Errorf("error slot type %s not found, %s", slotTypeName, err)
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %s: %s", slotTypeName, err)
		}

		return nil
	}
}

func testAccCheckAwsLexSlotTypeExists(slotTypeName string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return testAccCheckAwsLexSlotTypeExistsWithVersion(slotTypeName, LexSlotTypeVersionLatest, output)
}

func testAccCheckAwsLexSlotTypeNotExists(slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).lexmodelconn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})
		if isAWSErr(err, lexmodelbuildingservice.ErrCodeNotFoundException, "") {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %s: %s", slotTypeName, err)
		}

		return fmt.Errorf("error slot type exists %s", slotTypeName)
	}
}

func testAccAwsLexSlotTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
  create_version = true
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_withoutVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
  create_version = false
}
`, rName)
}

func testAccAwsLexSlotTypeUpdateConfig_description(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  description = "Types of flowers to pick up"
  name        = "%s"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_enumerationValues(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
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

  name = "%s"
}
`, rName)
}

func testAccAwsLexSlotTypeConfig_valueSelectionStrategy(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  name                     = "%s"
  value_selection_strategy = "TOP_RESOLUTION"
  enumeration_value {
    synonyms = [
      "Lirium",
      "Martagon",
    ]
    value = "lilies"
  }
}
`, rName)
}
