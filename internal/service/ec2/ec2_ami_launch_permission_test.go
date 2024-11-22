// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"context"
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccEC2AMILaunchPermission_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, names.AttrAccountID),
					resource.TestCheckResourceAttr(resourceName, "group", ""),
					resource.TestCheckResourceAttr(resourceName, "organization_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAMILaunchPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceAMILaunchPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_Disappears_ami(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tfec2.ResourceAMICopy(), "aws_ami_copy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_group(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_group(rName, "unblocked"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "group", "all"),
					resource.TestCheckResourceAttr(resourceName, "organization_arn", ""),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAMILaunchPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
			{
				Config: testAccAMILaunchPermissionConfig_group(rName, "block-new-sharing"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
				),
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_organizationARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsEnabled(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_organizationARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "group", ""),
					resource.TestCheckResourceAttrSet(resourceName, "organization_arn"),
					resource.TestCheckResourceAttr(resourceName, "organizational_unit_arn", ""),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAMILaunchPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_organizationalUnitARN(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckOrganizationsAccount(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckAMILaunchPermissionDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_organizationalUnitARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(ctx, resourceName),
					resource.TestCheckResourceAttr(resourceName, names.AttrAccountID, ""),
					resource.TestCheckResourceAttr(resourceName, "group", ""),
					resource.TestCheckResourceAttr(resourceName, "organization_arn", ""),
					resource.TestCheckResourceAttrSet(resourceName, "organizational_unit_arn"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateIdFunc: testAccAMILaunchPermissionImportStateIdFunc(resourceName),
				ImportStateVerify: true,
			},
		},
	})
}

func testAccAMILaunchPermissionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		imageID := rs.Primary.Attributes["image_id"]

		if v := rs.Primary.Attributes["group"]; v != "" {
			return fmt.Sprintf("%s/%s", v, imageID), nil
		} else if v := rs.Primary.Attributes["organization_arn"]; v != "" {
			return fmt.Sprintf("%s/%s", v, imageID), nil
		} else if v := rs.Primary.Attributes["organizational_unit_arn"]; v != "" {
			return fmt.Sprintf("%s/%s", v, imageID), nil
		} else {
			return fmt.Sprintf("%s/%s", rs.Primary.Attributes[names.AttrAccountID], imageID), nil
		}
	}
}

func testAccCheckAMILaunchPermissionExists(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		_, err := tfec2.FindImageLaunchPermission(ctx, conn, rs.Primary.Attributes["image_id"], rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["group"], rs.Primary.Attributes["organization_arn"], rs.Primary.Attributes["organizational_unit_arn"])

		return err
	}
}

func testAccCheckAMILaunchPermissionDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Client(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_ami_launch_permission" {
				continue
			}

			_, err := tfec2.FindImageLaunchPermission(ctx, conn, rs.Primary.Attributes["image_id"], rs.Primary.Attributes[names.AttrAccountID], rs.Primary.Attributes["group"], rs.Primary.Attributes["organization_arn"], rs.Primary.Attributes["organizational_unit_arn"])

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("AMI Launch Permission %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccAMILaunchPermissionConfig_accountID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  account_id = data.aws_caller_identity.current.account_id
  image_id   = aws_ami_copy.test.id
}
`, rName))
}

func testAccAMILaunchPermissionConfig_imagePublicAccess(state string) string {
	return fmt.Sprintf(`
resource "aws_ec2_image_block_public_access" "test" {
  state = %[1]q
}
`, state)
}

func testAccAMILaunchPermissionConfig_group(rName, state string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), testAccAMILaunchPermissionConfig_imagePublicAccess(state), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.name
  deprecation_time  = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.deprecation_time

  depends_on = [aws_ec2_image_block_public_access.test]
}

resource "aws_ami_launch_permission" "test" {
  group    = "all"
  image_id = aws_ami_copy.test.id
}
`, rName, state))
}

func testAccAMILaunchPermissionConfig_organizationARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  organization_arn = data.aws_organizations_organization.current.arn
  image_id         = aws_ami_copy.test.id
}
`, rName))
}

func testAccAMILaunchPermissionConfig_organizationalUnitARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinux2HVMEBSX8664AMI(), fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn2-ami-minimal-hvm-ebs-x86_64.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  organizational_unit_arn = aws_organizations_organizational_unit.test.arn
  image_id                = aws_ami_copy.test.id
}
`, rName))
}
