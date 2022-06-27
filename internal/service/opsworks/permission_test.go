package opsworks_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
)

func TestAccOpsWorksPermission_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_permission.test"
	var opsperm opsworks.Permission
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckPermissionDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_create(rName, true, true, "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					testAccCheckCreatePermissionAttributes(&opsperm, true, true, "iam_only"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "true"),
					resource.TestCheckResourceAttr(resourceName, "level", "iam_only"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, true, false, "iam_only"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					testAccCheckCreatePermissionAttributes(&opsperm, true, false, "iam_only"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "false"),
					resource.TestCheckResourceAttr(resourceName, "level", "iam_only"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, false, false, "deny"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					testAccCheckCreatePermissionAttributes(&opsperm, false, false, "deny"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "false"),
					resource.TestCheckResourceAttr(resourceName, "level", "deny"),
				),
			},
			{
				Config: testAccPermissionConfig_create(rName, false, false, "show"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					testAccCheckCreatePermissionAttributes(&opsperm, false, false, "show"),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "false"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "false"),
					resource.TestCheckResourceAttr(resourceName, "level", "show"),
				),
			},
		},
	})
}

// Reference: https://github.com/hashicorp/terraform-provider-aws/issues/4804
func TestAccOpsWorksPermission_self(t *testing.T) {
	var opsperm opsworks.Permission
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_opsworks_permission.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckPartitionHasService(opsworks.EndpointsID, t) },
		ErrorCheck:        acctest.ErrorCheck(t, opsworks.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      nil, // Cannot delete own OpsWorks Permission
		Steps: []resource.TestStep{
			{
				Config: testAccPermissionConfig_self(rName, true, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "true"),
				),
			},
			{
				Config: testAccPermissionConfig_self(rName, true, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckPermissionExists(resourceName, &opsperm),
					resource.TestCheckResourceAttr(resourceName, "allow_ssh", "true"),
					resource.TestCheckResourceAttr(resourceName, "allow_sudo", "false"),
				),
			},
		},
	})
}

func testAccCheckPermissionExists(
	n string, opsperm *opsworks.Permission) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

		params := &opsworks.DescribePermissionsInput{
			StackId:    aws.String(rs.Primary.Attributes["stack_id"]),
			IamUserArn: aws.String(rs.Primary.Attributes["user_arn"]),
		}
		resp, err := conn.DescribePermissions(params)

		if err != nil {
			return err
		}

		if v := len(resp.Permissions); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opsperm = *resp.Permissions[0]

		return nil
	}
}

func testAccCheckCreatePermissionAttributes(
	opsperm *opsworks.Permission, allowSSH bool, allowSudo bool, level string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsperm.AllowSsh != allowSSH {
			return fmt.Errorf("Unnexpected allowSSH: %t", *opsperm.AllowSsh)
		}

		if *opsperm.AllowSudo != allowSudo {
			return fmt.Errorf("Unnexpected allowSudo: %t", *opsperm.AllowSudo)
		}

		if *opsperm.Level != level {
			return fmt.Errorf("Unnexpected level: %s", *opsperm.Level)
		}

		return nil
	}
}

func testAccCheckPermissionDestroy(s *terraform.State) error {
	client := acctest.Provider.Meta().(*conns.AWSClient).OpsWorksConn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_permission" {
			continue
		}

		req := &opsworks.DescribePermissionsInput{
			IamUserArn: aws.String(rs.Primary.Attributes["user_arn"]),
		}

		resp, err := client.DescribePermissions(req)
		if err == nil {
			if len(resp.Permissions) > 0 {
				return fmt.Errorf("OpsWorks Permissions still exist.")
			}
		}

		if awserr, ok := err.(awserr.Error); ok {
			if awserr.Code() != "ResourceNotFoundException" {
				return err
			}
		}
	}
	return nil
}

func testAccPermissionBase(rName string) string {
	return fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  cidr_block = aws_vpc.test.cidr_block
  vpc_id     = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_opsworks_stack" "test" {
  name                         = %[1]q
  region                       = data.aws_region.current.name
  vpc_id                       = aws_vpc.test.id
  default_subnet_id            = aws_subnet.test.id
  service_role_arn             = aws_iam_role.service.arn
  default_instance_profile_arn = aws_iam_instance_profile.test.arn
  default_os                   = "Amazon Linux 2016.09"
  default_root_device_type     = "ebs"

  custom_json = <<EOF
{
  "key": "value"
}
EOF

  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "service" {
  name = "%[1]s-service"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "opsworks.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_role_policy" "service" {
  name = %[1]q
  role = aws_iam_role.service.id

  policy = <<EOT
{
  "Statement": [
    {
      "Action": [
        "ec2:*",
        "iam:PassRole",
        "cloudwatch:GetMetricStatistics",
        "elasticloadbalancing:*",
        "rds:*"
      ],
      "Effect": "Allow",
      "Resource": [
        "*"
      ]
    }
  ]
}
EOT
}

resource "aws_iam_role" "instance" {
  name = "%[1]s-instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "ec2.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
EOT
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = aws_iam_role.instance.name
}
`, rName)
}

func testAccPermissionConfig_create(rName string, allowSSH, allowSudo bool, level string) string {
	return acctest.ConfigCompose(
		testAccStackConfig_vpcCreate(rName),
		fmt.Sprintf(`
resource "aws_opsworks_permission" "test" {
  stack_id = aws_opsworks_stack.test.id

  allow_ssh  = %[1]t
  allow_sudo = %[2]t
  user_arn   = aws_opsworks_user_profile.user.user_arn
  level      = %[3]q
}

resource "aws_opsworks_user_profile" "user" {
  user_arn     = aws_iam_user.user.arn
  ssh_username = aws_iam_user.user.name
}

resource "aws_iam_user" "user" {
  name = %[4]q
  path = "/"
}
`, allowSSH, allowSudo, level, rName))
}

func testAccPermissionConfig_self(rName string, allowSSH bool, allowSudo bool) string {
	return acctest.ConfigCompose(
		testAccPermissionBase(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}

resource "aws_opsworks_permission" "test" {
  allow_ssh  = %[1]t
  allow_sudo = %[2]t
  stack_id   = aws_opsworks_stack.test.id
  user_arn   = data.aws_caller_identity.current.arn
}
`, allowSSH, allowSudo))
}
