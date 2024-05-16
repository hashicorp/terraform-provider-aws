// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThingGroup_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(resourceName, names.AttrARN, "iot", regexache.MustCompile(fmt.Sprintf("thinggroup/%s$", rName))),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", acctest.Ct1),
					resource.TestCheckResourceAttrSet(resourceName, "metadata.0.creation_date"),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.parent_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.root_to_parent_thing_groups.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, rName),
					resource.TestCheckResourceAttr(resourceName, "parent_group_name", ""),
					resource.TestCheckResourceAttr(resourceName, "properties.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
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

func TestAccIoTThingGroup_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceThingGroup(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingGroup_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingGroupConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
			{
				Config: testAccThingGroupConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
				),
			},
		},
	})
}

func TestAccIoTThingGroup_parentGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"
	parentResourceName := "aws_iot_thing_group.parent"
	grandparentResourceName := "aws_iot_thing_group.grandparent"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_parent(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(resourceName, "parent_group_name", parentResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "metadata.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.parent_group_name", parentResourceName, names.AttrName),
					resource.TestCheckResourceAttr(resourceName, "metadata.0.root_to_parent_groups.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.0.group_arn", grandparentResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.0.group_name", grandparentResourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.1.group_arn", parentResourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(resourceName, "metadata.0.root_to_parent_groups.1.group_name", parentResourceName, names.AttrName),
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

func TestAccIoTThingGroup_properties(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupConfig_properties(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.%", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key1", "Value1"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 1"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct1),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccThingGroupConfig_propertiesUpdated(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.%", acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key2", "Value2"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.attribute_payload.0.attributes.Key3", "Value3"),
					resource.TestCheckResourceAttr(resourceName, "properties.0.description", "test description 2"),
					resource.TestCheckResourceAttr(resourceName, names.AttrVersion, acctest.Ct2),
				),
			},
		},
	})
}

func testAccCheckThingGroupExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing Group ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindThingGroupByName(ctx, conn, rs.Primary.ID)

		return err
	}
}

func testAccCheckThingGroupDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_group" {
				continue
			}

			_, err := tfiot.FindThingGroupByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Thing Group %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccThingGroupConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q
}
`, rName)
}

func testAccThingGroupConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1)
}

func testAccThingGroupConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2)
}

func testAccThingGroupConfig_parent(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "grandparent" {
  name = "%[1]s-grandparent"
}

resource "aws_iot_thing_group" "parent" {
  name = "%[1]s-parent"

  parent_group_name = aws_iot_thing_group.grandparent.name
}

resource "aws_iot_thing_group" "test" {
  name = %[1]q

  parent_group_name = aws_iot_thing_group.parent.name
}
`, rName)
}

func testAccThingGroupConfig_properties(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  properties {
    attribute_payload {
      attributes = {
        Key1 = "Value1"
      }
    }

    description = "test description 1"
  }
}
`, rName)
}

func testAccThingGroupConfig_propertiesUpdated(rName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q

  properties {
    attribute_payload {
      attributes = {
        Key2 = "Value2"
        Key3 = "Value3"
      }
    }

    description = "test description 2"
  }
}
`, rName)
}
