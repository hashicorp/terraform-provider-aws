// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package lexmodels_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lexmodelbuildingservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tflexmodels "github.com/hashicorp/terraform-provider-aws/internal/service/lexmodels"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccLexModelsSlotType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					testAccCheckSlotTypeNotExists(ctx, testSlotTypeID, acctest.Ct1),
					resource.TestCheckResourceAttr(rName, "create_version", acctest.CtFalse),
					resource.TestCheckResourceAttr(rName, names.AttrDescription, ""),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", acctest.Ct1),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						names.AttrValue: "lilies",
					}),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Lirium"),
					resource.TestCheckTypeSetElemAttr(rName, "enumeration_value.*.synonyms.*", "Martagon"),
					resource.TestCheckResourceAttr(rName, names.AttrName, testSlotTypeID),
					resource.TestCheckResourceAttr(rName, "value_selection_strategy", lexmodelbuildingservice.SlotValueSelectionStrategyOriginalValue),
					resource.TestCheckResourceAttrSet(rName, "checksum"),
					resource.TestCheckResourceAttr(rName, names.AttrVersion, tflexmodels.SlotTypeVersionLatest),
					acctest.CheckResourceAttrRFC3339(rName, names.AttrCreatedDate),
					acctest.CheckResourceAttrRFC3339(rName, names.AttrLastUpdatedDate),
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

func TestAccLexModelsSlotType_createVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					testAccCheckSlotTypeNotExists(ctx, testSlotTypeID, acctest.Ct1),
					resource.TestCheckResourceAttr(rName, names.AttrVersion, tflexmodels.SlotTypeVersionLatest),
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
					testAccCheckSlotTypeExists(ctx, rName, &v),
					testAccCheckSlotTypeExistsWithVersion(ctx, rName, acctest.Ct1, &v),
					resource.TestCheckResourceAttr(rName, names.AttrVersion, acctest.Ct1),
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

func TestAccLexModelsSlotType_description(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, names.AttrDescription, ""),
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
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, names.AttrDescription, "Types of flowers to pick up"),
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

func TestAccLexModelsSlotType_enumerationValues(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", acctest.Ct1),
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
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, "enumeration_value.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(rName, "enumeration_value.*", map[string]string{
						names.AttrValue: "tulips",
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

func TestAccLexModelsSlotType_name(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID1 := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)
	testSlotTypeID2 := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID1),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, names.AttrName, testSlotTypeID1),
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
					testAccCheckSlotTypeExists(ctx, rName, &v),
					resource.TestCheckResourceAttr(rName, names.AttrName, testSlotTypeID2),
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

func TestAccLexModelsSlotType_valueSelectionStrategy(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
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
					testAccCheckSlotTypeExists(ctx, rName, &v),
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

func TestAccLexModelsSlotType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var v lexmodelbuildingservice.GetSlotTypeOutput
	rName := "aws_lex_slot_type.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccSlotTypeConfig_basic(testSlotTypeID),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSlotTypeExists(ctx, rName, &v),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tflexmodels.ResourceSlotType(), rName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccLexModelsSlotType_computeVersion(t *testing.T) {
	ctx := acctest.Context(t)
	var v1 lexmodelbuildingservice.GetSlotTypeOutput
	var v2 lexmodelbuildingservice.GetIntentOutput

	slotTypeResourceName := "aws_lex_slot_type.test"
	intentResourceName := "aws_lex_intent.test"
	testSlotTypeID := "test_slot_type_" + sdkacctest.RandStringFromCharSet(8, sdkacctest.CharSetAlpha)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, lexmodelbuildingservice.EndpointsID)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.LexModelsServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckSlotTypeDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeConfig_withVersion(testSlotTypeID),
					testAccIntentConfig_slotsWithVersion(testSlotTypeID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExistsWithVersion(ctx, slotTypeResourceName, acctest.Ct1, &v1),
					resource.TestCheckResourceAttr(slotTypeResourceName, names.AttrVersion, acctest.Ct1),
					testAccCheckIntentExistsWithVersion(ctx, intentResourceName, acctest.Ct1, &v2),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, acctest.Ct1),
					resource.TestCheckResourceAttr(intentResourceName, "slot.0.slot_type_version", acctest.Ct1),
				),
			},
			{
				Config: acctest.ConfigCompose(
					testAccSlotTypeUpdateConfig_enumerationValuesWithVersion(testSlotTypeID),
					testAccIntentConfig_slotsWithVersion(testSlotTypeID),
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckSlotTypeExistsWithVersion(ctx, slotTypeResourceName, acctest.Ct2, &v1),
					resource.TestCheckResourceAttr(slotTypeResourceName, names.AttrVersion, acctest.Ct2),
					resource.TestCheckResourceAttr(slotTypeResourceName, "enumeration_value.#", acctest.Ct2),
					resource.TestCheckTypeSetElemNestedAttrs(slotTypeResourceName, "enumeration_value.*", map[string]string{
						names.AttrValue: "tulips",
					}),
					resource.TestCheckTypeSetElemAttr(slotTypeResourceName, "enumeration_value.*.synonyms.*", "Eduardoregelia"),
					resource.TestCheckTypeSetElemAttr(slotTypeResourceName, "enumeration_value.*.synonyms.*", "Podonix"),
					testAccCheckIntentExistsWithVersion(ctx, intentResourceName, acctest.Ct2, &v2),
					resource.TestCheckResourceAttr(intentResourceName, names.AttrVersion, acctest.Ct2),
					resource.TestCheckResourceAttr(intentResourceName, "slot.0.slot_type_version", acctest.Ct2),
				),
			},
		},
	})
}

func testAccCheckSlotTypeExistsWithVersion(ctx context.Context, rName, slotTypeVersion string, v *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[rName]
		if !ok {
			return fmt.Errorf("Not found: %s", rName)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Lex Slot Type ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn(ctx)

		output, err := tflexmodels.FindSlotTypeVersionByName(ctx, conn, rs.Primary.ID, slotTypeVersion)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckSlotTypeExists(ctx context.Context, rName string, output *lexmodelbuildingservice.GetSlotTypeOutput) resource.TestCheckFunc {
	return testAccCheckSlotTypeExistsWithVersion(ctx, rName, tflexmodels.SlotTypeVersionLatest, output)
}

func testAccCheckSlotTypeNotExists(ctx context.Context, slotTypeName, slotTypeVersion string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn(ctx)

		_, err := tflexmodels.FindSlotTypeVersionByName(ctx, conn, slotTypeName, slotTypeVersion)

		if tfresource.NotFound(err) {
			return nil
		}

		if err != nil {
			return err
		}

		return fmt.Errorf("Lex Slot Type %s/%s still exists", slotTypeName, slotTypeVersion)
	}
}

func testAccCheckSlotTypeDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).LexModelsConn(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_lex_slot_type" {
				continue
			}

			output, err := conn.GetSlotTypeVersionsWithContext(ctx, &lexmodelbuildingservice.GetSlotTypeVersionsInput{
				Name: aws.String(rs.Primary.ID),
			})

			if err != nil {
				return err
			}

			if output == nil || len(output.SlotTypes) == 0 {
				return nil
			}

			return fmt.Errorf("Lex Slot Type %s still exists", rs.Primary.ID)
		}

		return nil
	}
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
