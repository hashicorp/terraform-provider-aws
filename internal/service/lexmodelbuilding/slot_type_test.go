package lexmodelbuilding_test

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	"github.com/hashicorp/aws-sdk-go-base/tfawserr"
	multierror "github.com/hashicorp/go-multierror"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflexmodelbuilding "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodelbuilding"
	"github.com/hashicorp/terraform-provider-aws/internal/sweep"
)





func TestAccLexModelBuildingSlotType_basic(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
					testAccCheckSlotTypeNotExists(testSlotTypeID, "1"),
					resource.TestCheckResourceAttr(rName, "create_version", "false"),
					resource.TestCheckResourceAttr(rName, "description", ""),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "1"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						"value": "lilies",
					}),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Lirium"),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Martagon"),
					resource.TestCheckResourceAttr(rName, "name", testSlotTypeID),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, "version", tflexmodelbuilding.SlotTypeVersionLatest),
					acctest.CheckResourceAttrRFC3339(rName, "created_date"),
					acctest.CheckResourceAttrRFC3339(rName, "last_updated_date"),
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

func TestAccLexModelBuildingSlotType_createVersion(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
					testAccCheckSlotTypeNotExists(testSlotTypeID, "1"),
					resource.TestCheckResourceAttr(rName, "version", tflexmodelbuilding.SlotTypeVersionLatest),
				),
			},
			{
				ResourceName:            rName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"create_version"},
			},
			{
				Config: testAccSlotTypeConfig_withVersion(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
					testAccCheckSlotTypeExistsWithVersion(rName, "1", &v),
					resource.TestCheckResourceAttr(rName, "version", "1"),
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

func TestAccLexModelBuildingSlotType_description(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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
				Config: testAccSlotTypeUpdateConfig_description(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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

func TestAccLexModelBuildingSlotType_enumerationValues(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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
				Config: testAccSlotTypeConfig_enumerationValues(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						"value": "tulips",
					}),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Eduardoregelia"),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Podonix"),
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

func TestAccLexModelBuildingSlotType_name(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID1 := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testSlotTypeID2 := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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
				Config: testAccSlotTypeConfig_basic(testSlotTypeID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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

func TestAccLexModelBuildingSlotType_valueSelectionStrategy(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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
				Config: testAccSlotTypeConfig_valueSelectionStrategy(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
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

func TestAccLexModelBuildingSlotType_disappears(t *testing.T) {
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(rName, &v),
					acctest.CheckResourceDisappears(acctest.Provider, tflexmodelbuilding.ResourceSlotType(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexModelBuildingSlotType_computeVersion(t *testing.T) {
	var v1 lexmodelbuildingservice.GetSlotTypeOutput
	var v2 lexmodelbuildingservice.GetIntentOutput

	slotTypeResourceName := "aws_lex_slot_type.test"
	intentResourceName := "aws_lex_intent.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	version := "1"
	updatedVersion := "2"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckPartitionHasService(lexmodelbuildingservice.EndpointsID, t)
		},
		ErrorCheck:   acctest.ErrorCheck(t, lexmodelbuildingservice.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckSlotTypeDestroy,
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_withVersion(testSlotTypeID),
					testAccIntentConfig_slotsWithVersion(testSlotTypeID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExistsWithVersion(slotTypeResourceName, version, &v1),
					resource.TestCheckResourceAttr(slotTypeResourceName, "version", version),
					testAccCheckIntentExistsWithVersion(intentResourceName, version, &v2),
					resource.TestCheckResourceAttr(intentResourceName, "version", version),
					resource.TestCheckResourceAttr(intentResourceName, "slot.0.slot_type_version", version),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeUpdateConfig_enumerationValuesWithVersion(testSlotTypeID),
					testAccIntentConfig_slotsWithVersion(testSlotTypeID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExistsWithVersion(slotTypeResourceName, updatedVersion, &v1),
					resource.TestCheckResourceAttr(slotTypeResourceName, "version", updatedVersion),
					resource.TestCheckResourceAttr(slotTypeResourceName, "enumeration_value.#", "2"),
					resource.TestCheckTypeSetElemNestedAttrs(slotTypeResourceName, "enumeration_value.*", map[string]string{
						"value": "tulips",
					}),
					resource.TestCheckTypeSetElemAttr(slotTypeResourceName, "enumeration_value.*.synonyms.*", "Eduardoregelia"),
					resource.TestCheckTypeSetElemAttr(slotTypeResourceName, "enumeration_value.*.synonyms.*", "Podonix"),
					testAccCheckIntentExistsWithVersion(intentResourceName, updatedVersion, &v2),
					resource.TestCheckResourceAttr(intentResourceName, "version", updatedVersion),
					resource.TestCheckResourceAttr(intentResourceName, "slot.0.slot_type_version", updatedVersion),
				),
			},
		},
	})
}

func testAccCheckSlotTypeExistsWithVersion(rName, slotTypeVersion string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex slot type ID is set")
		}

		var err error
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelBuildingConn

		output, err = conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(rs.Primary.ID),
			Version: aws.String(slotTypeVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return fmt.Errorf("error slot type %q version %s not found", rs.Primary.ID, slotTypeVersion)
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %q version %s: %w", rs.Primary.ID, slotTypeVersion, err)
		}

		return nil
	}
}

func testAccCheckSlotTypeExists(rName string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return testAccCheckSlotTypeExistsWithVersion(rName, tflexmodelbuilding.SlotTypeVersionLatest, output)
}

func testAccCheckSlotTypeNotExists(slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelBuildingConn

		_, err := conn.GetSlotType(&lexmodelbuildingservice.GetSlotTypeInput{
			Name:    aws.String(slotTypeName),
			Version: aws.String(slotTypeVersion),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			return nil
		}
		if err != nil {
			return fmt.Errorf("error getting slot type %s version %s: %s", slotTypeName, slotTypeVersion, err)
		}

		return fmt.Errorf("error slot type %s version %s exists", slotTypeName, slotTypeVersion)
	}
}

func testAccCheckSlotTypeDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelBuildingConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_lex_slot_type" {
			continue
		}

		output, err := conn.GetSlotTypeVersions(&lexmodelbuildingservice.GetSlotTypeVersionsInput{
			Name: aws.String(rs.Primary.ID),
		})
		if tfawserr.ErrCodeEquals(err, lexmodelbuildingservice.ErrCodeNotFoundException) {
			continue
		}
		if err != nil {
			return err
		}

		if output == nil || len(output.SlotTypes) == 0 {
			return nil
		}

		return fmt.Errorf("Lex slot type %q still exists", rs.Primary.ID)
	}

	return nil
}

func testAccSlotTypeConfig_basic(rName string) string {
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
}
`, rName)
}

func testAccSlotTypeConfig_withVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  create_version = true
  name           = "%s"
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

func testAccSlotTypeUpdateConfig_description(rName string) string {
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

func testAccSlotTypeConfig_enumerationValues(rName string) string {
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

  enumeration_value {
    synonyms = [
      "Eduardoregelia",
      "Podonix",
    ]
    value = "tulips"
  }
}
`, rName)
}

func testAccSlotTypeConfig_valueSelectionStrategy(rName string) string {
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

func testAccSlotTypeUpdateConfig_enumerationValuesWithVersion(rName string) string {
	return fmt.Sprintf(`
resource "aws_lex_slot_type" "test" {
  create_version = true
  name           = "%s"
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
}
`, rName)
}
