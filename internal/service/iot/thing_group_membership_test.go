// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package iot_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfiot "github.com/hashicorp/terraform-provider-aws/internal/service/iot"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccIoTThingGroupMembership_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupMembershipExists(ctx, resourceName),
					resource.TestCheckNoResourceAttr(resourceName, "override_dynamic_group"),
					resource.TestCheckResourceAttr(resourceName, "thing_group_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "thing_name", rName2),
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

func TestAccIoTThingGroupMembership_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupMembershipExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceThingGroupMembership(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingGroupMembership_disappears_Thing(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group_membership.test"
	thingResourceName := "aws_iot_thing.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupMembershipExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceThing(), thingResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingGroupMembership_disappears_ThingGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group_membership.test"
	thingGroupResourceName := "aws_iot_thing_group.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_basic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupMembershipExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfiot.ResourceThingGroup(), thingGroupResourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccIoTThingGroupMembership_overrideDynamicGroup(t *testing.T) {
	ctx := acctest.Context(t)
	rName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_iot_thing_group_membership.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.IoTServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckThingGroupMembershipDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccThingGroupMembershipConfig_overrideDynamic(rName1, rName2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckThingGroupMembershipExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "override_dynamic_group", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "thing_group_name", rName1),
					resource.TestCheckResourceAttr(resourceName, "thing_name", rName2),
				),
			},
			{
				ResourceName:            resourceName,
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"override_dynamic_group"},
			},
		},
	})
}

func testAccCheckThingGroupMembershipExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		_, err := tfiot.FindThingGroupMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing_group_name"], rs.Primary.Attributes["thing_name"])

		return err
	}
}

func testAccCheckThingGroupMembershipDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).IoTClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_iot_thing_group_membership" {
				continue
			}

			_, err := tfiot.FindThingGroupMembershipByTwoPartKey(ctx, conn, rs.Primary.Attributes["thing_group_name"], rs.Primary.Attributes["thing_name"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("IoT Thing Group Membership %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccThingGroupMembershipConfig_basic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q
}

resource "aws_iot_thing" "test" {
  name = %[2]q
}

resource "aws_iot_thing_group_membership" "test" {
  thing_group_name = aws_iot_thing_group.test.name
  thing_name       = aws_iot_thing.test.name
}
`, rName1, rName2)
}

func testAccThingGroupMembershipConfig_overrideDynamic(rName1, rName2 string) string {
	return fmt.Sprintf(`
resource "aws_iot_thing_group" "test" {
  name = %[1]q
}

resource "aws_iot_thing" "test" {
  name = %[2]q
}

resource "aws_iot_thing_group_membership" "test" {
  thing_group_name = aws_iot_thing_group.test.name
  thing_name       = aws_iot_thing.test.name

  override_dynamic_group = true
}
`, rName1, rName2)
}
