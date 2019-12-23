package aws

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/opsworks"
)

///////////////////////////////
//// Tests for the No-VPC case
///////////////////////////////

func TestAccAWSOpsworksStack_noVpcBasic(t *testing.T) {
	var opsstack opsworks.Stack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckHasDefaultVpcOrEc2Classic(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksStackConfigNoVpcCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &opsstack),
					testAccCheckAWSOpsworksCreateStackAttributes(&opsstack, rName),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						rName,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_os",
						"Amazon Linux 2016.09",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_root_device_type",
						"ebs",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_json",
						`{"key": "value"}`,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"configuration_manager_version",
						"11.10",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"use_opsworks_security_groups",
						"false",
					),
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

func TestAccAWSOpsworksStack_noVpcChangeServiceRoleForceNew(t *testing.T) {
	var before, after opsworks.Stack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckHasDefaultVpcOrEc2Classic(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksStackConfigNoVpcCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &before),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOpsworksStackConfigNoVpcCreateUpdateServiceRole(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &after),
					testAccCheckAWSOpsworksStackRecreated(t, &before, &after),
				),
			},
		},
	})
}

func TestAccAWSOpsworksStack_vpc(t *testing.T) {
	var opsstack opsworks.Stack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_stack.test"
	subnetResourceName := "aws_subnet.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksStackConfigVpcCreate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, true, &opsstack),
					testAccCheckAWSOpsworksCreateStackAttributes(&opsstack, rName),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						rName,
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"default_availability_zone",
						subnetResourceName,
						"availability_zone",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_os",
						"Amazon Linux 2016.09",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_root_device_type",
						"ebs",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_json",
						`{"key": "value"}`,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"configuration_manager_version",
						"11.10",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"use_opsworks_security_groups",
						"false",
					),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAWSOpsworksStackConfigVpcUpdate(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, true, &opsstack),
					testAccCheckAWSOpsworksUpdateStackAttributes(&opsstack, rName),
					resource.TestCheckResourceAttr(
						resourceName,
						"name",
						rName,
					),
					resource.TestCheckResourceAttrPair(
						resourceName,
						"default_availability_zone",
						subnetResourceName,
						"availability_zone",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_os",
						"Amazon Linux 2015.09",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"default_root_device_type",
						"ebs",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_json",
						`{"key": "value"}`,
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"configuration_manager_version",
						"11.10",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"use_opsworks_security_groups",
						"false",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"use_custom_cookbooks",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"manage_berkshelf",
						"true",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_cookbooks_source.0.type",
						"git",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_cookbooks_source.0.revision",
						"master",
					),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_cookbooks_source.0.url",
						"https://github.com/aws/opsworks-example-cookbooks.git",
					),
				),
			},
		},
	})
}

func TestAccAWSOpsworksStack_noVpcCreateTags(t *testing.T) {
	var opsstack opsworks.Stack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckHasDefaultVpcOrEc2Classic(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsworksStackConfigNoVpcCreateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &opsstack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.foo", "bar"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccAwsOpsworksStackConfigNoVpcUpdateTags(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &opsstack),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.wut", "asdf"),
				),
			},
		},
	})
}

/////////////////////////////
// Tests for Custom Cookbooks
/////////////////////////////

func TestAccAWSOpsworksStack_CustomCookbooks_SetPrivateProperties(t *testing.T) {
	var opsstack opsworks.Stack
	rName := acctest.RandomWithPrefix("tf-acc-test")
	resourceName := "aws_opsworks_stack.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSOpsworksStackConfig_CustomCookbooks_Set(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, true, &opsstack),
					testAccCheckAWSOpsworksCreateStackAttributesWithCookbooks(&opsstack, rName),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_cookbooks_source.0.password",
						"password"),
					resource.TestCheckResourceAttr(
						resourceName,
						"custom_cookbooks_source.0.ssh_key",
						sshKey),
				),
			},
		},
	})
}

// Tests the addition of regional endpoints and supporting the classic link used
// to create Stack's prior to v0.9.0.
// See https://github.com/hashicorp/terraform/issues/12842
func TestAccAWSOpsworksStack_classicEndpoints(t *testing.T) {
	stackName := fmt.Sprintf("tf-opsworks-acc-%d", acctest.RandInt())
	resourceName := "aws_opsworks_stack.main"
	rInt := acctest.RandInt()
	var opsstack opsworks.Stack

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsOpsworksStackDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsOpsWorksStack_classic_endpoint(stackName, rInt),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSOpsworksStackExists(resourceName, false, &opsstack),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			// Ensure that changing to us-west-2 region results in no plan
			{
				Config:   testAccAwsOpsWorksStack_regional_endpoint(stackName, rInt),
				PlanOnly: true,
			},
		},
	})
}

func testAccCheckAWSOpsworksStackRecreated(t *testing.T,
	before, after *opsworks.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *before.StackId == *after.StackId {
			t.Fatalf("Expected change of Opsworks StackIds, but both were %v", before.StackId)
		}
		return nil
	}
}

////////////////////////////
//// Checkers and Utilities
////////////////////////////

func testAccCheckAwsOpsworksStackDestroy(s *terraform.State) error {
	opsworksconn := testAccProvider.Meta().(*AWSClient).opsworksconn
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_opsworks_stack" {
			continue
		}

		req := &opsworks.DescribeStacksInput{
			StackIds: []*string{
				aws.String(rs.Primary.ID),
			},
		}

		_, err := opsworksconn.DescribeStacks(req)
		if err != nil {
			if awserr, ok := err.(awserr.Error); ok {
				if awserr.Code() == "ResourceNotFoundException" {
					// not found, all good
					return nil
				}
			}
			return err
		}
	}
	return fmt.Errorf("Fall through error for OpsWorks stack test")
}

func testAccCheckAWSOpsworksStackExists(
	n string, vpc bool, opsstack *opsworks.Stack) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).opsworksconn

		params := &opsworks.DescribeStacksInput{
			StackIds: []*string{aws.String(rs.Primary.ID)},
		}
		resp, err := conn.DescribeStacks(params)

		if err != nil {
			return err
		}

		if v := len(resp.Stacks); v != 1 {
			return fmt.Errorf("Expected 1 response returned, got %d", v)
		}

		*opsstack = *resp.Stacks[0]

		if vpc {
			if rs.Primary.Attributes["vpc_id"] != *opsstack.VpcId {
				return fmt.Errorf("VPCID Got %s, expected %s", *opsstack.VpcId, rs.Primary.Attributes["vpc_id"])
			}
			if rs.Primary.Attributes["default_subnet_id"] != *opsstack.DefaultSubnetId {
				return fmt.Errorf("Default subnet Id Got %s, expected %s", *opsstack.DefaultSubnetId, rs.Primary.Attributes["default_subnet_id"])
			}
		}

		return nil
	}
}

func testAccCheckAWSOpsworksCreateStackAttributes(opsstack *opsworks.Stack, stackName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsstack.Name != stackName {
			return fmt.Errorf("Unexpected stackName: %s", *opsstack.Name)
		}

		if *opsstack.DefaultOs != "Amazon Linux 2016.09" {
			return fmt.Errorf("Unexpected DefaultOs: %s", *opsstack.DefaultOs)
		}

		if *opsstack.DefaultRootDeviceType != "ebs" {
			return fmt.Errorf("Unexpected DefaultRootDeviceType: %s", *opsstack.DefaultRootDeviceType)
		}

		if *opsstack.CustomJson != `{"key": "value"}` {
			return fmt.Errorf("Unexpected CustomJson: %s", *opsstack.CustomJson)
		}

		if *opsstack.ConfigurationManager.Version != "11.10" {
			return fmt.Errorf("Unexpected ConfigurationManager.Version: %s", *opsstack.ConfigurationManager.Version)
		}

		if *opsstack.UseOpsworksSecurityGroups {
			return fmt.Errorf("Unexpected UseOpsworksSecurityGroups: %t", *opsstack.UseOpsworksSecurityGroups)
		}

		return nil
	}
}

func testAccCheckAWSOpsworksCreateStackAttributesWithCookbooks(opsstack *opsworks.Stack, stackName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsstack.Name != stackName {
			return fmt.Errorf("Unexpected stackName: %s", *opsstack.Name)
		}

		if *opsstack.DefaultOs != "Amazon Linux 2016.09" {
			return fmt.Errorf("Unexpected DefaultOs: %s", *opsstack.DefaultOs)
		}

		if *opsstack.DefaultRootDeviceType != "ebs" {
			return fmt.Errorf("Unexpected DefaultRootDeviceType: %s", *opsstack.DefaultRootDeviceType)
		}

		if *opsstack.CustomJson != `{"key": "value"}` {
			return fmt.Errorf("Unexpected CustomJson: %s", *opsstack.CustomJson)
		}

		if *opsstack.ConfigurationManager.Version != "11.10" {
			return fmt.Errorf("Unexpected ConfigurationManager.Version: %s", *opsstack.ConfigurationManager.Version)
		}

		if *opsstack.UseOpsworksSecurityGroups {
			return fmt.Errorf("Unexpected UseOpsworksSecurityGroups: %t", *opsstack.UseOpsworksSecurityGroups)
		}

		if !*opsstack.UseCustomCookbooks {
			return fmt.Errorf("Unexpected UseCustomCookbooks: %t", *opsstack.UseCustomCookbooks)
		}

		if *opsstack.CustomCookbooksSource.Type != "git" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Type)
		}

		if *opsstack.CustomCookbooksSource.Revision != "master" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Revision)
		}

		if *opsstack.CustomCookbooksSource.Url != "https://github.com/aws/opsworks-example-cookbooks.git" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Url)
		}

		if *opsstack.CustomCookbooksSource.Username != "username" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Username)
		}

		return nil
	}
}

func testAccCheckAWSOpsworksUpdateStackAttributes(opsstack *opsworks.Stack, stackName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if *opsstack.Name != stackName {
			return fmt.Errorf("Unexpected stackName: %s", *opsstack.Name)
		}

		if *opsstack.DefaultOs != "Amazon Linux 2015.09" {
			return fmt.Errorf("Unexpected DefaultOs: %s", *opsstack.DefaultOs)
		}

		if *opsstack.DefaultRootDeviceType != "ebs" {
			return fmt.Errorf("Unexpected DefaultRootDeviceType: %s", *opsstack.DefaultRootDeviceType)
		}

		if *opsstack.CustomJson != `{"key": "value"}` {
			return fmt.Errorf("Unexpected CustomJson: %s", *opsstack.CustomJson)
		}

		if *opsstack.ConfigurationManager.Version != "11.10" {
			return fmt.Errorf("Unexpected ConfigurationManager.Version: %s", *opsstack.ConfigurationManager.Version)
		}

		if !*opsstack.UseCustomCookbooks {
			return fmt.Errorf("Unexpected UseCustomCookbooks: %t", *opsstack.UseCustomCookbooks)
		}

		if !*opsstack.ChefConfiguration.ManageBerkshelf {
			return fmt.Errorf("Unexpected ChefConfiguration.ManageBerkshelf: %t", *opsstack.ChefConfiguration.ManageBerkshelf)
		}

		if *opsstack.CustomCookbooksSource.Type != "git" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Type)
		}

		if *opsstack.CustomCookbooksSource.Revision != "master" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Revision)
		}

		if *opsstack.CustomCookbooksSource.Url != "https://github.com/aws/opsworks-example-cookbooks.git" {
			return fmt.Errorf("Unexpected CustomCookbooksSource.Type: %s", *opsstack.CustomCookbooksSource.Url)
		}

		return nil
	}
}

func testAccAwsOpsWorksStack_classic_endpoint(rName string, rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-east-1"
}

resource "aws_opsworks_stack" "main" {
  name                         = "%s"
  region                       = "us-west-2"
  service_role_arn             = "${aws_iam_role.opsworks_service.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.opsworks_instance.arn}"

  configuration_manager_version = "12"
  default_availability_zone     = "us-west-2b"
}

resource "aws_iam_role" "opsworks_service" {
  name = "tf_opsworks_service_%d"

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

resource "aws_iam_role_policy" "opsworks_service" {
  name = "tf_opsworks_service_%d"
  role = "${aws_iam_role.opsworks_service.id}"

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
      "Resource": ["*"]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "tf_opsworks_instance_%d"

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

resource "aws_iam_instance_profile" "opsworks_instance" {
  name  = "%s_profile"
  roles = ["${aws_iam_role.opsworks_instance.name}"]
}
`, rName, rInt, rInt, rInt, rName)
}

func testAccAwsOpsWorksStack_regional_endpoint(rName string, rInt int) string {
	return fmt.Sprintf(`
provider "aws" {
  region = "us-west-2"
}

resource "aws_opsworks_stack" "main" {
  name                         = "%s"
  region                       = "us-west-2"
  service_role_arn             = "${aws_iam_role.opsworks_service.arn}"
  default_instance_profile_arn = "${aws_iam_instance_profile.opsworks_instance.arn}"

  configuration_manager_version = "12"
  default_availability_zone     = "us-west-2b"
}

resource "aws_iam_role" "opsworks_service" {
  name = "tf_opsworks_service_%d"

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

resource "aws_iam_role_policy" "opsworks_service" {
  name = "tf_opsworks_service_%d"
  role = "${aws_iam_role.opsworks_service.id}"

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
      "Resource": ["*"]
    }
  ]
}
EOT
}

resource "aws_iam_role" "opsworks_instance" {
  name = "tf_opsworks_instance_%d"

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

resource "aws_iam_instance_profile" "opsworks_instance" {
  name  = "%s_profile"
  roles = ["${aws_iam_role.opsworks_instance.name}"]
}
`, rName, rInt, rInt, rInt, rName)
}

func testAccAwsOpsworksStackConfigNoVpcCreate(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "current" {
  state = "available"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_availability_zone     = "${data.aws_availability_zones.current.names[0]}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}
`, rName)
}

func testAccAwsOpsworksStackConfigNoVpcCreateTags(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "current" {
  state = "available"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_availability_zone     = "${data.aws_availability_zones.current.names[0]}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false

  tags = {
    foo = "bar"
  }
}
`, rName)
}

func testAccAwsOpsworksStackConfigNoVpcUpdateTags(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "current" {
  state = "available"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_availability_zone     = "${data.aws_availability_zones.current.names[0]}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false

  tags = {
    wut = "asdf"
  }
}
`, rName)
}

func testAccAwsOpsworksStackConfigNoVpcCreateUpdateServiceRole(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

data "aws_availability_zones" "current" {
  state = "available"
}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  service_role_arn              = "${aws_iam_role.test2.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_availability_zone     = "${data.aws_availability_zones.current.names[0]}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}

resource "aws_iam_role" "test2" {
  name = "%[1]s_2"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [{
    "Sid": "",
    "Effect": "Allow",
    "Principal": {"Service": "opsworks.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy" "test2" {
  name = "%[1]s_2"
  role = "${aws_iam_role.test2.id}"

  policy = <<EOT
{
  "Statement": [{
    "Action": [
      "ec2:*",
      "iam:PassRole",
      "cloudwatch:*",
      "elasticloadbalancing:*",
      "rds:*"
    ],
    "Effect": "Allow",
    "Resource": ["*"]
  }]
}
EOT
}
`, rName)
}

////////////////////////////
//// Tests for the VPC case
////////////////////////////

func testAccAwsOpsworksStackConfigVpcCreate(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + testAccAWSOpsworksStackConfig_vpcResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  vpc_id                        = "${aws_vpc.test.id}"
  default_subnet_id             = "${aws_subnet.test.id}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
}
`, rName)
}

func testAccAWSOpsworksStackConfigVpcUpdate(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + testAccAWSOpsworksStackConfig_vpcResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  vpc_id                        = "${aws_vpc.test.id}"
  default_subnet_id             = "${aws_subnet.test.id}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_os                    = "Amazon Linux 2015.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
  use_custom_cookbooks          = true
  manage_berkshelf              = true

  custom_cookbooks_source {
    type     = "git"
    revision = "master"
    url      = "https://github.com/aws/opsworks-example-cookbooks.git"
  }
}
`, rName)
}

/////////////////////////////////////////
// Helpers for Custom Cookbook properties
/////////////////////////////////////////

func testAccAWSOpsworksStackConfig_CustomCookbooks_Set(rName string) string {
	return testAccAWSOpsworksStackConfig_iamResources(rName) + testAccAWSOpsworksStackConfig_vpcResources(rName) + fmt.Sprintf(`
data "aws_region" "current" {}

resource "aws_opsworks_stack" "test" {
  name                          = %[1]q
  region                        = "${data.aws_region.current.name}"
  vpc_id                        = "${aws_vpc.test.id}"
  default_subnet_id             = "${aws_subnet.test.id}"
  service_role_arn              = "${aws_iam_role.test.arn}"
  default_instance_profile_arn  = "${aws_iam_instance_profile.test.arn}"
  default_os                    = "Amazon Linux 2016.09"
  default_root_device_type      = "ebs"
  custom_json                   = "{\"key\": \"value\"}"
  configuration_manager_version = "11.10"
  use_opsworks_security_groups  = false
  use_custom_cookbooks          = true
  manage_berkshelf              = true

  custom_cookbooks_source {
    type     = "git"
    revision = "master"
    url      = "https://github.com/aws/opsworks-example-cookbooks.git"
    username = "username"
    password = "password"
    ssh_key  = %[2]q
  }
}
`, rName, sshKey)
}

func testAccAWSOpsworksStackConfig_iamResources(rName string) string {
	return fmt.Sprintf(`
resource "aws_iam_role" "test" {
  name = %[1]q

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [{
    "Sid": "",
    "Effect": "Allow",
    "Principal": {"Service": "opsworks.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_role_policy" "test" {
  name = %[1]q
  role = "${aws_iam_role.test.id}"

  policy = <<EOT
{
  "Statement": [{
    "Action": [
      "ec2:*",
      "iam:PassRole",
      "cloudwatch:GetMetricStatistics",
      "elasticloadbalancing:*",
      "rds:*"
    ],
    "Effect": "Allow",
    "Resource": ["*"]
  }]
}
EOT
}

resource "aws_iam_role" "test_instance" {
  name = "%[1]s_instance"

  assume_role_policy = <<EOT
{
  "Version": "2008-10-17",
  "Statement": [{
    "Sid": "",
    "Effect": "Allow",
    "Principal": {"Service": "ec2.amazonaws.com"},
    "Action": "sts:AssumeRole"
  }]
}
EOT
}

resource "aws_iam_instance_profile" "test" {
  name = %[1]q
  role = "${aws_iam_role.test_instance.name}"
}
`, rName)
}

func testAccAWSOpsworksStackConfig_vpcResources(rName string) string {
	return fmt.Sprintf(`
data "aws_availability_zones" "current" {
  state = "available"
}

resource "aws_vpc" "test" {
  cidr_block = "10.3.5.0/24"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "${aws_vpc.test.cidr_block}"
  availability_zone = "${data.aws_availability_zones.current.names[0]}"

  tags = {
    Name = %[1]q
  }
}
`, rName)
}

// One-off, bogus private key generated for use in testing
const sshKey = "-----BEGIN RSA PRIVATE KEY-----" +
	"MIIEpAIBAAKCAQEAv/1hnOZadSDMbJUVJsqweDwc/4TvhTGf0vl9vtNyjzqUUxgU" +
	"RrSvYrgkvWgAFtQ9J5QDNOPSRvS8F1cu7tR036cecdHPmA+Cxto1qENy8UeYrKzV" +
	"I55i+vJiSn3i22HW+SbW1raBM+PL3sp9i0BQmCr8eh3i/VdUm92OQHtnjhfLB3GX" +
	"xnrvytBfI8p2bx9j7mAAjS/X+QncMawPqI9WGuizmuC2cTQHZpZY7j/w+bItoYIV" +
	"g5qJV3908LNlNZGU6etdEUTWM1VSNxG2Yk6eULeStSA4oSkJSHlwP1/fjab0j1b4" +
	"HeB/TUFpy3ODrRAhuHxlyFFWMSzePkXLx9d0GwIDAQABAoIBAFlwrj/M5Ik6XWGc" +
	"Vj07IdjxkETNZlQzmRRNHHKAyRbGoIDRb+i8lhQ0WxFN2PTJrS+5+YBzPevGabWp" +
	"7PhgS45BqaI2rzJUz4TZ9TNNMMgMpaiT37t3Nv9XWckAOmYff2mU2XMvlKNa1QgW" +
	"Z0QvExzAsdwl/jAttgHixjluBAEib+G3p0Xt2CZMQYNzE9H2gH/nqkysiZ5fC+ng" +
	"RnM843jAHtrfz9Q0ATBADMJZgZepnMZyldaOV+s5L8UB893UGhrfGrBwlHd5U5ug" +
	"Z/p74IvOgDd3/pp/2yuyqE+RWz9sakss196aJ0jUXVXjH3F+QDdqqPx0YIJ7S0eM" +
	"13T7hGkCgYEA4TqpoPFYIVEug4gQ6SDttSMwrbA5uBM13s1vjNpDBFuHWFMqgSRe" +
	"xlIAGCGNhoyTr3xr/34filwGMkMdLw8JFISOIbZ18+qgDOsSW0tXwE03vQgKFNB1" +
	"ClGEfcd/4B/oLwOe/bqnKVBQSnfp05yqHjdc9XNQeFxLL8LfSv7LIIUCgYEA2jgt" +
	"108LF+RtdkmSoqLnexJ0jmbPlcYTw1wzuIg49gLdlRxoC+UwPFc+uzMGNxEzr6hG" +
	"Eg3dJVr3+TMLIcTD6usPWzzuL4ReV/IAhCjzgS/WopqURg4cQ+R4MjvTMg8GCZfE" +
	"QvjcbpKh5ndP/QQEOy7cAP8BLVSG3/ichMcttB8CgYAdzmaebvILzrOKIpqiT4JF" +
	"w3dwtO6ehqRNbQCDMmtGC1rY/ICWgJquQjHS/7W8BaSRx7R/JlDEPbNwOWOGU8YO" +
	"2g/5NC1d70HpE77lKA5f25gxwvuaj4+9otYW0y0AGxjeB+ulhmsS05cck8v0/jmh" +
	"MBB0RyNyGjy1AGQOh7OYBQKBgQCwFq1HFM2K1hVOYkglXPcV5OqRDn1sCo5gEsLZ" +
	"oXL1cZKEhIuhLawixPQl8yKMxSDEGjGQ2Acf4axANuRAt5qwskWOBjjdtx66MNoh" +
	"yznTgVrdk4cakMBWOMKVJplhx6XDj+gbct3NjB2A775oGRmg+Esnsp6siYzcpq0G" +
	"qANFWQKBgQCyv8KoQXsD8f8XMvicRC42uZXfhlDjOUzpo1O7WQKWBYqPBqz4AHzE" +
	"Cdy6djI120bqDOifre1qnBjoHezrG+ejaQOTpocOVwT5Zl7BhjoXQZRGiQXj+2aD" +
	"tmm0+hpmkjX7jiPcljjs8S8gh+uCWieJoO4JNPk2SXRiePpYgKzdlg==" +
	"-----END RSA PRIVATE KEY-----"
