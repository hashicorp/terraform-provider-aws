package ec2_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMILaunchPermission(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// Bug reference: https://github.com/hashicorp/terraform-provider-aws/issues/6222.
// Images with <group>all</group> will not have <userId> and can cause a panic.
func TestAccEC2AMILaunchPermission_DisappearsLaunchPermission_public(t *testing.T) {
	resourceName := "aws_ami_launch_permission.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					testAccCheckAMILaunchPermissionAddPublic(resourceName),
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
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckAMILaunchPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAMILaunchPermissionConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAMILaunchPermissionExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tfec2.ResourceAMICopy(), "aws_ami_copy.test"),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
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

		imageID, accountID, err := tfec2.AMILaunchPermissionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		_, err = tfec2.FindImageLaunchPermission(context.TODO(), conn, imageID, accountID)

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

		imageID, accountID, err := tfec2.AMILaunchPermissionParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		_, err = tfec2.FindImageLaunchPermission(context.TODO(), conn, imageID, accountID)

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

func testAccCheckAMILaunchPermissionAddPublic(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).EC2Conn

		input := &ec2.ModifyImageAttributeInput{
			Attribute: aws.String(ec2.ImageAttributeNameLaunchPermission),
			ImageId:   aws.String(rs.Primary.Attributes["image_id"]),
			LaunchPermission: &ec2.LaunchPermissionModifications{
				Add: []*ec2.LaunchPermission{
					{Group: aws.String(ec2.PermissionGroupAll)},
				},
			},
		}

		_, err := conn.ModifyImageAttribute(input)

		if err != nil {
			return err
		}

		return err
	}
}

func testAccAMILaunchPermissionConfig(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigLatestAmazonLinuxHvmEbsAmi(), fmt.Sprintf(`
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

func testAccAMILaunchPermissionImportStateIdFunc(resourceName string) resource.ImportStateIdFunc {
	return func(s *terraform.State) (string, error) {
		rs, ok := s.RootModule().Resources[resourceName]
		if !ok {
			return "", fmt.Errorf("Not found: %s", resourceName)
		}

		return fmt.Sprintf("%s/%s", rs.Primary.Attributes["account_id"], rs.Primary.Attributes["image_id"]), nil
	}
}
