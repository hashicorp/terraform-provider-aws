// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamFleetStackAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_fleet_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, "fleet_name", rName),
					resource.TestCheckResourceAttr(resourceName, "stack_name", rName),
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

func TestAccAppStreamFleetStackAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_fleet_stack_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetStackAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceFleetStackAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetStackAssociationExists(ctx context.Context, resourceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return fmt.Errorf("not found: %s", resourceName)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		fleetName, stackName, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
		if err != nil {
			return fmt.Errorf("error decoding AppStream Fleet Stack Association ID (%s): %w", rs.Primary.ID, err)
		}

		err = tfappstream.FindFleetStackAssociation(ctx, conn, fleetName, stackName)

		if tfresource.NotFound(err) {
			return fmt.Errorf("AppStream Fleet Stack Association %q does not exist", rs.Primary.ID)
		}

		return err
	}
}

func testAccCheckFleetStackAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_fleet_stack_association" {
				continue
			}

			fleetName, stackName, err := tfappstream.DecodeStackFleetID(rs.Primary.ID)
			if err != nil {
				return fmt.Errorf("error decoding AppStream Fleet Stack Association ID (%s): %w", rs.Primary.ID, err)
			}

			err = tfappstream.FindFleetStackAssociation(ctx, conn, fleetName, stackName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Fleet Stack Association %q still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccFleetStackAssociationConfig_basic(name string) string {
	// "Amazon-AppStream2-Sample-Image-03-11-2023" is not available in GovCloud
	return fmt.Sprintf(`
resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  image_name    = "Amazon-AppStream2-Sample-Image-03-11-2023"
  instance_type = "stream.standard.small"

  compute_capacity {
    desired_instances = 1
  }
}

resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_fleet_stack_association" "test" {
  fleet_name = aws_appstream_fleet.test.name
  stack_name = aws_appstream_stack.test.name
}
`, name)
}
