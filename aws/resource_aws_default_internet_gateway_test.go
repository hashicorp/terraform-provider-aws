package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsDefaultInternetGateway_basic(t *testing.T) {
	var igw ec2.InternetGateway

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsDefaultInternetGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAwsDefaultInternetGatewayConfig,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckInternetGatewayExists("aws_default_internet_gateway.foo", &igw),
					resource.TestCheckResourceAttr(
						"aws_default_internet_gateway.foo", "tags.%", "1"),
					resource.TestCheckResourceAttr(
						"aws_default_internet_gateway.foo", "tags.Name", "terraform-testacc-default-igw"),
				),
			},
		},
	})
}

func testAccCheckAwsDefaultInternetGatewayDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).ec2conn

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_default_internet_gateway" {
			continue
		}

		// Try to find the resource
		resp, err := conn.DescribeInternetGateways(&ec2.DescribeInternetGatewaysInput{
			InternetGatewayIds: []*string{aws.String(rs.Primary.ID)},
		})
		if err != nil {
			return err
		}
		if len(resp.InternetGateways) != 1 {
			return fmt.Errorf("does not exist")
		}

		return nil
	}

	return nil
}

const testAccAwsDefaultInternetGatewayConfig = `
resource "aws_default_internet_gateway" "foo" {
  tags {
    Name = "terraform-testacc-default-igw"
  }
}
`
