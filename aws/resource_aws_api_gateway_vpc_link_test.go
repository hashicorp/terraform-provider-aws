package aws

import (
	"fmt"
	"log"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigateway"
	"github.com/hashicorp/terraform/helper/acctest"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func init() {
	resource.AddTestSweepers("aws_api_gateway_vpc_link", &resource.Sweeper{
		Name: "aws_api_gateway_vpc_link",
		F:    testSweepAPIGatewayVpcLinks,
	})
}

func testSweepAPIGatewayVpcLinks(region string) error {
	client, err := sharedClientForRegion(region)
	if err != nil {
		return fmt.Errorf("error getting client: %s", err)
	}
	conn := client.(*AWSClient).apigateway

	err = conn.GetVpcLinksPages(&apigateway.GetVpcLinksInput{}, func(page *apigateway.GetVpcLinksOutput, lastPage bool) bool {
		for _, item := range page.Items {
			input := &apigateway.DeleteVpcLinkInput{
				VpcLinkId: item.Id,
			}
			id := aws.StringValue(item.Id)

			log.Printf("[INFO] Deleting API Gateway VPC Link: %s", id)
			_, err := conn.DeleteVpcLink(input)

			if err != nil {
				log.Printf("[ERROR] Failed to delete API Gateway VPC Link %s: %s", id, err)
				continue
			}

			if err := waitForApiGatewayVpcLinkDeletion(conn, id); err != nil {
				log.Printf("[ERROR] Error waiting for API Gateway VPC Link (%s) deletion: %s", id, err)
			}
		}
		return !lastPage
	})
	if err != nil {
		if testSweepSkipSweepError(err) {
			log.Printf("[WARN] Skipping API Gateway VPC Link sweep for %s: %s", region, err)
			return nil
		}
		return fmt.Errorf("Error retrieving API Gateway VPC Links: %s", err)
	}

	return nil
}

func TestAccAWSAPIGatewayVpcLink_basic(t *testing.T) {
	rName := acctest.RandString(5)
	resource.ParallelTest(t, resource.TestCase{
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
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "target_arns.#", "1"),
				),
			},
			{
				Config: testAccAPIGatewayVpcLinkConfig_Update(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckAwsAPIGatewayVpcLinkExists("aws_api_gateway_vpc_link.test"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "name", fmt.Sprintf("tf-apigateway-update-%s", rName)),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "description", "test update"),
					resource.TestCheckResourceAttr("aws_api_gateway_vpc_link.test", "target_arns.#", "1"),
				),
			},
		},
	})
}

func TestAccAWSAPIGatewayVpcLink_importBasic(t *testing.T) {
	rName := acctest.RandString(5)
	resourceName := "aws_api_gateway_vpc_link.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAwsAPIGatewayVpcLinkDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAPIGatewayVpcLinkConfig(rName),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
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

		_, err := conn.GetVpcLink(input)
		if err != nil {
			if isAWSErr(err, apigateway.ErrCodeNotFoundException, "") {
				return nil
			}
			return err
		}

		return fmt.Errorf("Expected VPC Link to be destroyed, %s found", rs.Primary.ID)
	}

	return nil
}

func testAccCheckAwsAPIGatewayVpcLinkExists(name string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[name]
		if !ok {
			return fmt.Errorf("Not found: %s", name)
		}

		conn := testAccProvider.Meta().(*AWSClient).apigateway

		input := &apigateway.GetVpcLinkInput{
			VpcLinkId: aws.String(rs.Primary.ID),
		}

		_, err := conn.GetVpcLink(input)
		return err
	}
}

func testAccAPIGatewayVpcLinkConfig_basis(rName string) string {
	return fmt.Sprintf(`
resource "aws_lb" "test_a" {
  name               = "tf-lb-%s"
  internal           = true
  load_balancer_type = "network"
  subnets            = ["${aws_subnet.test.id}"]
}

resource "aws_vpc" "test" {
  cidr_block = "10.10.0.0/16"
}

data "aws_availability_zones" "test" {}

resource "aws_subnet" "test" {
  vpc_id            = "${aws_vpc.test.id}"
  cidr_block        = "10.10.0.0/21"
  availability_zone = "${data.aws_availability_zones.test.names[0]}"

  tags = {
    Name = "tf-acc-api-gateway-vpc-link"
  }
}
`, rName)
}

func testAccAPIGatewayVpcLinkConfig(rName string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-%s"
  description = "test"
  target_arns = ["${aws_lb.test_a.arn}"]
}
`, rName)
}

func testAccAPIGatewayVpcLinkConfig_Update(rName string) string {
	return testAccAPIGatewayVpcLinkConfig_basis(rName) + fmt.Sprintf(`
resource "aws_api_gateway_vpc_link" "test" {
  name = "tf-apigateway-update-%s"
  description = "test update"
  target_arns = ["${aws_lb.test_a.arn}"]
}
`, rName)
}
