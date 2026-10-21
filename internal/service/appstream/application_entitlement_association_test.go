// Copyright IBM Corp. 2014, 2025
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
	"github.com/hashicorp/terraform-provider-aws/internal/retry"
	tfappstream "github.com/hashicorp/terraform-provider-aws/internal/service/appstream"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccAppStreamApplicationEntitlementAssociation_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_application_entitlement_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationEntitlementAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationEntitlementAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationEntitlementAssociationExists(ctx, resourceName),
					resource.TestCheckResourceAttrSet(resourceName, "stack_name"),
					resource.TestCheckResourceAttrSet(resourceName, "entitlement_name"),
					resource.TestCheckResourceAttrSet(resourceName, "application_identifier"),
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

func TestAccAppStreamApplicationEntitlementAssociation_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_appstream_application_entitlement_association.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckApplicationEntitlementAssociationDestroy(ctx),
		ErrorCheck:               acctest.ErrorCheck(t, names.AppStreamServiceID),
		Steps: []resource.TestStep{
			{
				Config: testAccApplicationEntitlementAssociationConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckApplicationEntitlementAssociationExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfappstream.ResourceApplicationEntitlementAssociation(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccCheckApplicationEntitlementAssociationDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_appstream_application_entitlement_association" {
				continue
			}

			stackName := rs.Primary.Attributes["stack_name"]
			entitlementName := rs.Primary.Attributes["entitlement_name"]
			applicationIdentifier := rs.Primary.Attributes["application_identifier"]
			err := tfappstream.FindApplicationEntitlementAssociationByThreePartKey(ctx, conn, stackName, entitlementName, applicationIdentifier)

			if retry.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AppStream Application Entitlement Association %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckApplicationEntitlementAssociationExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).AppStreamClient(ctx)

		stackName := rs.Primary.Attributes["stack_name"]
		entitlementName := rs.Primary.Attributes["entitlement_name"]
		applicationIdentifier := rs.Primary.Attributes["application_identifier"]
		err := tfappstream.FindApplicationEntitlementAssociationByThreePartKey(ctx, conn, stackName, entitlementName, applicationIdentifier)

		return err
	}
}

func testAccApplicationEntitlementAssociationConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_appstream_stack" "test" {
  name = %[1]q
}

resource "aws_appstream_entitlement" "test" {
  name           = %[1]q
  stack_name     = aws_appstream_stack.test.name
  app_visibility = "ASSOCIATED"

  attributes {
    name  = "department"
    value = "engineering"
  }
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_object" "test" {
  bucket  = aws_s3_bucket.test.id
  key     = "application"
  content = "application content"
}

resource "aws_appstream_application" "test" {
  name                   = %[1]q
  display_name           = %[1]q
  icon_s3_location {
    s3_bucket = aws_s3_bucket.test.id
    s3_key    = aws_s3_object.test.key
  }
  launch_path            = "C:\\Program Files\\TestApp\\test.exe"
  platforms              = ["WINDOWS"]
  app_block_arn          = aws_appstream_app_block.test.arn
}

resource "aws_appstream_app_block" "test" {
  name                   = %[1]q
  source_s3_location {
    s3_bucket = aws_s3_bucket.test.id
    s3_key    = aws_s3_object.test.key
  }
}

resource "aws_appstream_application_fleet_association" "test" {
  application_arn = aws_appstream_application.test.arn
  fleet_name      = aws_appstream_fleet.test.name
}

resource "aws_appstream_fleet" "test" {
  name          = %[1]q
  instance_type = "stream.standard.small"
  compute_capacity {
    desired_instances = 1
  }
}

resource "aws_appstream_fleet_stack_association" "test" {
  fleet_name = aws_appstream_fleet.test.name
  stack_name = aws_appstream_stack.test.name
}

resource "aws_appstream_application_entitlement_association" "test" {
  stack_name             = aws_appstream_stack.test.name
  entitlement_name       = aws_appstream_entitlement.test.name
  application_identifier = aws_appstream_application.test.id

  depends_on = [
    aws_appstream_application_fleet_association.test,
    aws_appstream_fleet_stack_association.test
  ]
}
`, rName)
}
