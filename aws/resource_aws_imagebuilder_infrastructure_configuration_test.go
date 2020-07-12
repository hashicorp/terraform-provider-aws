package aws

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/imagebuilder"
	"github.com/hashicorp/terraform-plugin-sdk/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/terraform"
)

//Check the minimal configuration
func TestAccAWSImageBuilderInfrastructureConfiguration_basic(t *testing.T) {
	var conf imagebuilder.InfrastructureConfiguration
	RandomString := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"
	resource.ParallelTest(t, resource.TestCase{
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSImageBuilderInfrastructureConfiguration_basic(RandomString),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSImageBuilderInfrastructureConfigurationExist(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "instance_profile_name"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
				),
			},
		},
	})
}

//Check the configuration with everything
func TestAccAWSImageBuilderInfrastructureConfiguration_advanced(t *testing.T) {
	var conf imagebuilder.InfrastructureConfiguration
	RandomString := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSImageBuilderInfrastructureConfiguration_advanced(RandomString),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSImageBuilderInfrastructureConfigurationExist(resourceName, &conf),
					resource.TestCheckResourceAttr(resourceName, "instance_profile_name", RandomString),
					resource.TestCheckResourceAttr(resourceName, "name", RandomString),
					resource.TestCheckResourceAttr(resourceName, "description", "example desc"),
					resource.TestCheckResourceAttr(resourceName, "key_pair", RandomString),
					resource.TestCheckResourceAttrSet(resourceName, "sns_topic_arn"),
					resource.TestCheckResourceAttrSet(resourceName, "subnet_id"),
					resource.TestCheckResourceAttr(resourceName, "terminate_instance_on_failure", "false"),
					resource.TestCheckResourceAttr(resourceName, "security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "instance_types.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_bucket_name", RandomString),
					resource.TestCheckResourceAttr(resourceName, "logging.0.s3_logs.0.s3_key_prefix", "logs"),
				),
			},
		},
	})
}

//Test the tags work
func TestAccAWSImageBuilderInfrastructureConfiguration_tags(t *testing.T) {
	var conf imagebuilder.InfrastructureConfiguration
	RandomString := fmt.Sprintf("tf-test-%d", acctest.RandInt())
	resourceName := "aws_imagebuilder_infrastructure_configuration.test"

	resource.ParallelTest(t, resource.TestCase{
		IDRefreshName: resourceName,
		Providers:     testAccProviders,
		CheckDestroy:  testAccCheckAWSImageBuilderInfrastructureConfigurationDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSImageBuilderInfrastructureConfiguration_tags1(RandomString),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSImageBuilderInfrastructureConfigurationExist(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "instance_profile_name"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
				),
			},
			{
				Config: testAccAWSImageBuilderInfrastructureConfiguration_tags2(RandomString),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckAWSImageBuilderInfrastructureConfigurationExist(resourceName, &conf),
					resource.TestCheckResourceAttrSet(resourceName, "instance_profile_name"),
					resource.TestCheckResourceAttrSet(resourceName, "name"),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
				),
			},
		},
	})
}

func testAccCheckAWSImageBuilderInfrastructureConfigurationExist(n string, res *imagebuilder.InfrastructureConfiguration) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return errors.New("No ImageBuilderInfrastructureConfiguration ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		resp, err := conn.GetInfrastructureConfiguration(&imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err != nil {
			return err
		}

		if resp == nil || resp.InfrastructureConfiguration == nil {
			return fmt.Errorf("Infrastructure configuration (%s) not found", rs.Primary.Attributes["name"])
		}
		*res = *resp.InfrastructureConfiguration
		return nil
	}
}

func testAccCheckAWSImageBuilderInfrastructureConfigurationDestroy(s *terraform.State) error {

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_imagebuilder_infrastructure_configuration" {
			continue
		}

		conn := testAccProvider.Meta().(*AWSClient).imagebuilderconn

		resp, err := conn.GetInfrastructureConfiguration(&imagebuilder.GetInfrastructureConfigurationInput{
			InfrastructureConfigurationArn: aws.String(rs.Primary.Attributes["arn"]),
		})

		if err == nil {
			if resp.InfrastructureConfiguration != nil {
				return fmt.Errorf("ImageBuilderInfrastructureConfiguration %q still exists", rs.Primary.ID)
			}
		}

		// Verify the error
		if awsErr, ok := err.(awserr.Error); ok {
			if awsErr.Code() == "ResourceNotFoundException" {
				return nil
			}
		}
		return err
	}

	return nil
}

func testAccAWSImageBuilderInfrastructureConfiguration_basic(ConfigurationName string) string {
	return fmt.Sprintf(`

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test_profile.name
  name = "%[1]s"
}

resource "aws_iam_instance_profile" "test_profile" {
  name = "%[1]s"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name = "%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}
`, ConfigurationName)
}

func testAccAWSImageBuilderInfrastructureConfiguration_advanced(ConfigurationName string) string {
	return fmt.Sprintf(`
resource "aws_key_pair" "test" {
  key_name   = "%[1]s"
  public_key = "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABAQD3F6tyPEFEzV0LX3X8BsXdMsQz1x2cEikKDEY0aIj41qgxMCP/iteneqXSIFZBp5vizPvaoIR3Um9xK7PGoW8giupGn+EPuxIA4cDM4vzOqOkiMPhz5XK0whEjkVzTo4+S0puvDZuwIsdiW9mxhJc7tgBNL0cYlWSYVkz4G/fslNfRPW5mYAM49f4fhtxPb5ok4Q2Lg9dPKVHO/Bgeu5woMc7RY0p1ej6D4CKFE6lymSDJpW0YHX/wqE9+cfEauh7xZcG0q9t2ta6F6fmX0agvpFyZo8aFbXeUBr7osSCJNgvavWbM/06niWrOvYX2xwWdhXmXSrbX8ZbabVohBK41 example@example.com"
}

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test_profile.name
  name = "%[1]s"
  description = "example desc"
  instance_types = ["t3.micro"]
  key_pair = aws_key_pair.test.key_name
  logging  {
    s3_logs  {
      s3_bucket_name = aws_s3_bucket.example.bucket
      s3_key_prefix = "logs"
    }
  }
  security_group_ids = [aws_security_group.example.id]
  sns_topic_arn = aws_sns_topic.example.arn
  subnet_id = aws_subnet.main.id
  terminate_instance_on_failure = false
}

resource "aws_iam_instance_profile" "test_profile" {
  name = "%[1]s"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name = "%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}

resource "aws_sns_topic" "example" {
  name = "%[1]s"
}


resource "aws_subnet" "main" {
  vpc_id     = aws_vpc.main.id
  cidr_block = "10.0.1.0/24"
}


resource "aws_vpc" "main" {
  cidr_block = "10.0.0.0/16"
}

resource "aws_security_group" "example" {
  name        = "%[1]s"
  vpc_id      = aws_vpc.main.id
}

resource "aws_s3_bucket" "example" {
  bucket = %[1]q
  acl    = "private"
}

`, ConfigurationName)
}

func testAccAWSImageBuilderInfrastructureConfiguration_tags1(ConfigurationName string) string {
	return fmt.Sprintf(`

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test_profile.name
  name = "%[1]s"
  tags = {
    Name = "tf-test",
  }
}

resource "aws_iam_instance_profile" "test_profile" {
  name = "%[1]s"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name = "%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}
`, ConfigurationName)
}

func testAccAWSImageBuilderInfrastructureConfiguration_tags2(ConfigurationName string) string {
	return fmt.Sprintf(`

resource "aws_imagebuilder_infrastructure_configuration" "test" {
  instance_profile_name = aws_iam_instance_profile.test_profile.name
  name = "%[1]s"
  tags = {
    Name = "tf-test",
    ExtraName = "tf-test"
  }
}

resource "aws_iam_instance_profile" "test_profile" {
  name = "%[1]s"
  role = aws_iam_role.role.name
}

resource "aws_iam_role" "role" {
  name = "%[1]s"
  path = "/"

  assume_role_policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
       {
            "Action": "sts:AssumeRole",
            "Principal": {
               "Service": "ec2.amazonaws.com"
            },
            "Effect": "Allow",
            "Sid": ""
        }
    ]
}
EOF
}
`, ConfigurationName)
}
