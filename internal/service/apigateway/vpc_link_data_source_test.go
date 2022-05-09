package apigateway_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/apigateway"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccAPIGatewayVPCLinkDataSource_basic(t *testing.T) {
	rName := fmt.Sprintf("tf-acc-test-%s", sdkacctest.RandString(8))
	resourceName := "aws_api_gateway_vpc_link.vpc_link"
	dataSourceName := "data.aws_api_gateway_vpc_link.vpc_link"
	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, apigateway.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCLinkDataSourceConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "description", resourceName, "description"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status_message"),
					resource.TestCheckResourceAttrSet(dataSourceName, "status"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "target_arns.#", "1"),
				),
			},
		},
	})
}

func testAccVPCLinkDataSourceConfig(r string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "apigateway_vpclink_test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "terraform-testacc-lb-apigateway-vpc-link"
  }
}

resource "aws_lb" "apigateway_vpclink_test" {
  name = "%s"

  subnets = [
    aws_subnet.apigateway_vpclink_test_subnet1.id,
  ]

  load_balancer_type               = "network"
  internal                         = true
  idle_timeout                     = 60
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = false

  tags = {
    Name = "testAccDataSourceAwsApiGatewayVpcLinkConfig_networkLoadbalancer"
  }
}

resource "aws_lb" "apigateway_vpclink_test2" {
  name = "%s-wrong"

  subnets = [
    aws_subnet.apigateway_vpclink_test_subnet1.id,
  ]

  load_balancer_type               = "network"
  internal                         = true
  idle_timeout                     = 60
  enable_deletion_protection       = false
  enable_cross_zone_load_balancing = false

  tags = {
    Name = "testAccDataSourceAwsApiGatewayVpcLinkConfig_networkLoadbalancer"
  }
}

resource "aws_subnet" "apigateway_vpclink_test_subnet1" {
  vpc_id     = aws_vpc.apigateway_vpclink_test.id
  cidr_block = "10.0.1.0/24"

  tags = {
    Name = "tf-acc-lb-apigateway-vpclink"
  }
}

resource "aws_api_gateway_vpc_link" "vpc_link" {
  name        = "%s"
  target_arns = [aws_lb.apigateway_vpclink_test.arn]
}

resource "aws_api_gateway_vpc_link" "vpc_link2" {
  name        = "%s-wrong"
  target_arns = [aws_lb.apigateway_vpclink_test2.arn]
}

data "aws_api_gateway_vpc_link" "vpc_link" {
  name = aws_api_gateway_vpc_link.vpc_link.name
}
`, r, r, r, r)
}
