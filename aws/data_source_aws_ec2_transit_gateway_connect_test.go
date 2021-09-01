package aws

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func TestAccAWSEc2TransitGatewayConnectDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect.test"
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectDataSourceConfigFilter(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "protocol", dataSourceName, "protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", dataSourceName, "transport_transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func TestAccAWSEc2TransitGatewayConnectDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect.test"
	resourceName := "aws_ec2_transit_gateway_connect.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t); testAccPreCheckAWSEc2TransitGateway(t) },
		ErrorCheck:   testAccErrorCheck(t, ec2.EndpointsID),
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckAWSEc2TransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccAWSEc2TransitGatewayConnectDataSourceConfigID(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "protocol", dataSourceName, "protocol"),
					resource.TestCheckResourceAttrPair(resourceName, "transport_transit_gateway_attachment_id", dataSourceName, "transport_transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
				),
			},
		},
	})
}

func testAccAWSEc2TransitGatewayConnectDataSourceConfigFilter() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {

}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

data "aws_ec2_transit_gateway_connect" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_connect.test.id]
  }
}
`
}

func testAccAWSEc2TransitGatewayConnectDataSourceConfigID() string {
	return testAccAvailableAZsNoOptInDefaultExcludeConfig() + `
# IncorrectState: Transit Gateway is not available in availability zone usw2-az4	

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = "tf-acc-test-ec2-transit-gateway-connect"
  }
}

resource "aws_ec2_transit_gateway" "test" {
	
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id
}

resource "aws_ec2_transit_gateway_connect" "test" {
	transport_transit_gateway_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id
}

data "aws_ec2_transit_gateway_connect" "test" {
  id = aws_ec2_transit_gateway_connect.test.id
}
`
}
