// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThingType_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					acctest.CheckResourceAttrRegionalARNFormat(ctx, resourceName, names.AttrARN, "iot", "thingtype/{name}"),
					resource.TestCheckResourceAttr(resourceName, "deprecated", acctest.CtFalse),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrID, resourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccIoTThingType_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfiot.ResourceThingType(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingType_full(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_full(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deprecated", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "MyDescription"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.searchable_attributes.#", "3"),
					resource.TestCheckTypeSetElemAttr(resourceName, "properties.0.searchable_attributes.*", "foo"),
					resource.TestCheckTypeSetElemAttr(resourceName, "properties.0.searchable_attributes.*", "bar"),
					resource.TestCheckTypeSetElemAttr(resourceName, "properties.0.searchable_attributes.*", "baz"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingTypeConfig_full(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, "deprecated", acctest.CtFalse),
				),
			},
		},
	})
}

func TestAccIoTThingType_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_type.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingTypeDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccThingTypeConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingTypeConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "2"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccThingTypeConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingTypeExists(ctx, t, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, "1"),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func testAccCheckThingTypeExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		_, err := tfiot.FindThingTypeByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckThingTypeDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_type" {
				continue
			}

			_, err := tfiot.FindThingTypeByName(ctx, conn, rs.Primary.ID)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Thing Type %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccThingTypeConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "test" {
  name = %[1]q
}
`, rName)
}

func testAccThingTypeConfig_full(rName string, deprecated bool) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "test" {
  name       = %[1]q
  deprecated = %[2]t

  properties {
    description           = "MyDescription"
    searchable_attributes = ["foo", "bar", "baz"]
  }
}
`, rName, deprecated)
}

func testAccThingTypeConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "test" {
  name       = %[1]q
  deprecated = false

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccThingTypeConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_type" "test" {
  name       = %[1]q
  deprecated = false

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}
