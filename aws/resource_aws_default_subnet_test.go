package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDefaultSubnet_basic(t *testing.T) {
	var v ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDefaultSubnetConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.Name", "Default subnet for us-west-2a"),
				),
			},
		},
	})
}

func TestAccAwsDefaultSubnet_createNew(t *testing.T) {
	var v ec2.Subnet

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDefaultSubnetDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDefaultSubnetConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "map_public_ip_on_launch", "true"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "assign_ipv6_address_on_creation", "false"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "tags.Name", "Default subnet for us-west-2a"),
					testAccDeleteAwsDefaultSubnet(&v),
				),
				ExpectNonEmptyPlan: true,
			},
			{
				Config: testAccAwsDefaultSubnetConfigBasic,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckSubnetExists("aws_default_subnet.foo", &v),
					resource.TestCheckResourceAttr(
						"aws_default_subnet.foo", "availability_zone", "us-west-2a"),
				),
			},
		},
	})
}

func testAccCheckAwsDefaultSubnetDestroy(s *terraform.State) error {
	// We expect subnet to still exist
	return nil
}

func testAccDeleteAwsDefaultSubnet(sn *ec2.Subnet) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := testAccProvider.Meta().(*AWSClient).ec2conn

		_, err := conn.DeleteSubnet(&ec2.DeleteSubnetInput{
			SubnetId: sn.SubnetId,
		})
		if err != nil {
			return err
		}

		return nil
	}
}

const testAccAwsDefaultSubnetConfigBasic = `
provider "aws" {
  region = "us-west-2"
}

resource "aws_default_subnet" "foo" {
  availability_zone = "us-west-2a"
  tags {
    Name = "Default subnet for us-west-2a"
  }
}
`
