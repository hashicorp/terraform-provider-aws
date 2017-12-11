package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAWSLaunchTemplate_basic(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_basic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "default_version", "1"),
					resource.TestCheckResourceAttr(resName, "latest_version", "1"),
				),
			},
		},
	})
}

func TestAccAWSLaunchTemplate_data(t *testing.T) {
	var template ec2.LaunchTemplate
	resName := "aws_launch_template.foo"

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSLaunchTemplateDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSLaunchTemplateConfig_data,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAWSLaunchTemplateExists(resName, &template),
					resource.TestCheckResourceAttr(resName, "block_device_mappings.#", "1"),
					resource.TestCheckResourceAttr(resName, "credit_specification.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "disable_api_termination"),
					resource.TestCheckResourceAttr(resName, "elastic_gpu_specifications.#", "1"),
					resource.TestCheckResourceAttr(resName, "iam_instance_profile.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "image_id"),
					resource.TestCheckResourceAttrSet(resName, "instance_initiated_shutdown_behavior"),
					resource.TestCheckResourceAttr(resName, "instance_market_options.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "instance_type"),
					resource.TestCheckResourceAttrSet(resName, "kernel_id"),
					resource.TestCheckResourceAttrSet(resName, "key_name"),
					resource.TestCheckResourceAttrSet(resName, "monitoring"),
					resource.TestCheckResourceAttr(resName, "network_interfaces.#", "1"),
					resource.TestCheckResourceAttr(resName, "placement.#", "1"),
					resource.TestCheckResourceAttrSet(resName, "ram_disk_id"),
					resource.TestCheckResourceAttr(resName, "vpc_security_group_ids.#", "1"),
					resource.TestCheckResourceAttr(resName, "tag_specifications.#", "1"),
					resource.TestCheckResourceAttr(resName, "tags.#", "1"),
				),
			},
		},
	})
}

func testAccCheckAWSLaunchTemplateExists(n string, t *ec2.LaunchTemplate) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No Launch Template ID is set")
		}

		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}

		if len(resp.LaunchTemplates) != 1 || *resp.LaunchTemplates[0].LaunchTemplateId != rs.Primary.ID {
			return fmt.Errorf("Launch Template not found")
		}

		*t = *resp.LaunchTemplates[0]

		return nil
	}
}

func testAccCheckAWSLaunchTemplateDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_launch_template" {
			continue
		}

		resp, err := conn.DescribeLaunchTemplates(&ec2.DescribeLaunchTemplatesInput{
			LaunchTemplateIds: []*string{aws.String(rs.Primary.ID)},
		})

		if err == nil {
			if len(resp.LaunchTemplates) != 0 && *resp.LaunchTemplates[0].LaunchTemplateId == rs.Primary.ID {
				return fmt.Errorf("Launch Template still exists")
			}
		}

		ae, ok := err.(awserr.Error)
		if !ok {
			return err
		}
		if ae.Code() != "InvalidLaunchTemplateId.NotFound" {
			log.Printf("aws error code: %s", ae.Code())
			return err
		}
	}

	return nil
}

const testAccAWSLaunchTemplateConfig_basic = `
resource "aws_launch_template" "foo" {
  name = "foo"
}
`

const testAccAWSLaunchTemplateConfig_data = `
resource "aws_launch_template" "foo" {
  name = "foo"

  block_device_mappings {
    device_name = "test"
  }

  credit_specification {
    cpu_credits = "standard"
  }

  disable_api_termination = true

  ebs_optimized = true

  elastic_gpu_specifications {
    type = "test"
  }

  iam_instance_profile {
    name = "test"
  }

  image_id = "ami-test"

  instance_initiated_shutdown_behavior = "test"

  instance_market_options {
    market_type = "test"
  }

  instance_type = "t2.micro"

  kernel_id = "test"

  key_name = "test"

  monitoring = true

  network_interfaces {
    associate_public_ip_address = true
  }

  placement {
    availability_zone = "test"
  }

  ram_disk_id = "test"

  vpc_security_group_ids = ["test"]

  tag_specifications {
    resource_type = "instance"
    tags {
      Name = "test"
    }
  }
}
`
