package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tfec2 "github.com/hashicorp/terraform-provider-aws/internal/service/ec2"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
)

func TestAccEC2AMILaunchPermission_basic(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					acctest.CheckResourceAttrAccountID(resourceName, "account_id"),
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
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMILaunchPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_Disappears_ami(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_accountID(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMICopy(), "aws_ami_copy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccEC2AMILaunchPermission_group(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_group(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
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
		},
	})
}

func TestAccEC2AMILaunchPermission_organizationARN(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsEnabled(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_organizationARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
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
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckOrganizationsAccount(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig_organizationalUnitARN(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "account_id", ""),
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
			return fmt.Sprintf("%s/%s", rs.Primary.Attributes["account_id"], imageID), nil
		}
	}
}

func testAccCheckAMILaunchPermissionExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No AMI Launch Permission ID is set")
		}

		imageID, accountID, group, organizationARN, organizationalUnitARN, err := tfec2.AMILaunchPermissionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err = tfec2.FindImageLaunchPermission(context.TODO(), conn, imageID, accountID, group, organizationARN, organizationalUnitARN)

		if err != nil {
			return err
		}

		return nil
	}
}

func testAccCheckAMILaunchPermissionDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_ami_launch_permission" {
			continue
		}

		imageID, accountID, group, organizationARN, organizationalUnitARN, err := tfec2.AMILaunchPermissionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindImageLaunchPermission(context.TODO(), conn, imageID, accountID, group, organizationARN, organizationalUnitARN)

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

func testAccAMILaunchPermissionConfig_accountID(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
data "aws_caller_identity" "current" {}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  account_id = data.aws_caller_identity.current.account_id
  image_id   = aws_ami_copy.test.id
}
`, rName))
}

func testAccAMILaunchPermissionConfig_group(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  source_ami_region = data.aws_region.current.name
  deprecation_time  = data.aws_ami.amzn-ami-minimal-hvm-ebs.deprecation_time
}

resource "aws_ami_launch_permission" "test" {
  group    = "all"
  image_id = aws_ami_copy.test.id
}
`, rName))
}

func testAccAMILaunchPermissionConfig_organizationARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
data "aws_organizations_organization" "current" {}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  organization_arn = data.aws_organizations_organization.current.arn
  image_id         = aws_ami_copy.test.id
}
`, rName))
}

func testAccAMILaunchPermissionConfig_organizationalUnitARN(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHVMEBSAMI(), fmt.Sprintf(`
resource "aws_organizations_organization" "test" {}

resource "aws_organizations_organizational_unit" "test" {
  name      = %[1]q
  parent_id = aws_organizations_organization.test.roots[0].id
}

data "aws_region" "current" {}

resource "aws_ami_copy" "test" {
  description       = %[1]q
  name              = %[1]q
  source_ami_id     = data.aws_ami.amzn-ami-minimal-hvm-ebs.id
  source_ami_region = data.aws_region.current.name
}

resource "aws_ami_launch_permission" "test" {
  organizational_unit_arn = aws_organizations_organizational_unit.test.arn
  image_id                = aws_ami_copy.test.id
}
`, rName))
}
