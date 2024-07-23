// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iot"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThing_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingConfig_basic(thingName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(ctx, resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
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

func TestAccIoTThing_full(t *testing.T) {
	ctx := acctest.Context(t)
	var thing iot.DescribeThingOutput
	rString := sdkacctest.RandString(8)
	thingName := fmt.Sprintf("tf_acc_thing_%s", rString)
	typeName := fmt.Sprintf("tf_acc_type_%s", rString)
	resourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingConfig_full(thingName, typeName, "42"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(ctx, resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "42"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{ // Update attribute
				Config: testAccThingConfig_full(thingName, typeName, "differentOne"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(ctx, resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, thingName),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", typeName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct3),
					resource.TestCheckResourceAttr(resourceName, "attributes.One", "11111"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Two", "TwoTwo"),
					resource.TestCheckResourceAttr(resourceName, "attributes.Answer", "differentOne"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
			{ // Remove thing type association
				Config: testAccThingConfig_fullUpdated(thingName, typeName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingExists(ctx, resourceName, &thing),
					resource.TestCheckResourceAttr(resourceName, names.AttrName, thingName),
					resource.TestCheckResourceAttr(resourceName, "attributes.%", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "thing_type_name", ""),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(resourceName, "default_client_id"),
					resource.TestCheckResourceAttrSet(resourceName, names.AttrVersion),
				),
			},
		},
	})
}

func testAccCheckThingExists(ctx context.Context, n string, v *iot.DescribeThingOutput) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No IoT Thing ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		output, err := tfiot.FindThingByName(ctx, conn, rs.Primary.ID)

		if err != nil {
			return err
		}

		*v = *output

		return nil
	}
}

func testAccCheckThingDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing" {
				continue
			}

			_, err := tfiot.FindThingByName(ctx, conn, rs.Primary.ID)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Thing %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccThingConfig_basic(thingName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"
}
`, thingName)
}

func testAccThingConfig_full(thingName, typeName, answer string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = "%s"

  attributes = {
    One    = "11111"
    Two    = "TwoTwo"
    Answer = "%s"
  }

  thing_type_name = aws_iot_thing_type.test.name
}

resource "aws_iot_thing_type" "test" {
  name = "%s"
}
`, thingName, answer, typeName)
}

func testAccThingConfig_fullUpdated(thingName, typeName string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing" "test" {
  name = %[1]q
}

resource "aws_iot_thing_type" "test" {
  name = %[2]q
}
`, thingName, typeName)
}
