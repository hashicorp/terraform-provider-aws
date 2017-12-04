package aws

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccAwsAPIGatewayVpcLink_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists("aws_api_gateway_vpc_link.test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "name", fmt.Sprintf("tf-apigateway-%s", rName)),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "description", "test"),
				),
			},
			{
				Config: testAccAPIGatewayVpcLinkConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists("aws_api_gateway_vpc_link.test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "name", fmt.Sprintf("tf-apigateway-update-%s", rName)),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "description", "test update"),
				),
			},
		},
	})
}

func testAccCheckAwsAPIGatewayVpcLinkDestroy(s *terraform.State) error {
	conn := testAccProvider.Meta().(*AWSClient).apigateway

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_api_gateway_vpc_link" {
			continue
		}

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		resp, err := conn.GetVpcLink(input)
		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		if *resp.Status != apigateway.VpcLinkStatusDeleting {
			return fmt.Errorf("APIGateway VPC Link (%s) not deleted", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckAwsAPIGatewayVpcLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		_, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		return nil
	}
}

func testAccAPIGatewayVpcLinkConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name = "tf-lb-a-%s"
  internal = true
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id = "${aws_subnet.test.id}"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone = "us-west-2a"
}

resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-%s"
  description = "test"
  target_arn = "${aws_lb.test_a.arn}"
}
`, rName, rName)
}

func testAccAPIGatewayVpcLinkConfig_Update(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name = "tf-lb-a-%s"
  internal = true
  load_balancer_type = "network"

  subnet_mapping {
    subnet_id = "${aws_subnet.test.id}"
  }
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
}

resource "aws_subnet" "test" {
  vpc_id = "${aws_vpc.test.id}"
  cidr_block = "10.10.0.0/21"
  map_public_ip_on_launch = true
  availability_zone = "us-west-2a"
}

resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-update-%s"
  description = "test update"
  target_arn = "${aws_lb.test_a.arn}"
}
`, rName, rName)
}
