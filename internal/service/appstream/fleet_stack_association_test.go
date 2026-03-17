// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package appstream_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamFleetStackAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_fleet_stack_association.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetStackAssociationDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(ctx, t, resourceName),
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
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckHasIAMRole(ctx, t, "AmazonAppStreamServiceAccess")
		},
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckFleetStackAssociationDestroy(ctx, t),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccFleetStackAssociationConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckFleetStackAssociationExists(ctx, t, resourceName),
					acctest.CheckSDKResourceDisappears(ctx, t, tfappstream.ResourceFleetStackAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckFleetStackAssociationDestroy(ctx context.Context, t *testing.T) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_fleet_stack_association" {
				continue
			}

			err := tfappstream.FindFleetStackAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["fleet_name"], rs.Primary.Attributes["stack_name"])

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Fleet Stack Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckFleetStackAssociationExists(ctx context.Context, t *testing.T, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.ProviderMeta(ctx, t).AppStreamClient(ctx)

		err := tfappstream.FindFleetStackAssociationByTwoPartKey(ctx, conn, rs.Primary.Attributes["fleet_name"], rs.Primary.Attributes["stack_name"])

		return err
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
