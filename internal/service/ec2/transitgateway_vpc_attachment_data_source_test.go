package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayVPCAttachmentDataSource_Filter(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_vpc_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentDataSourceConfig_filter(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "appliance_mode_support", dataSourceName, "appliance_mode_support"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_support", dataSourceName, "ipv6_support"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", dataSourceName, "vpc_owner_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentDataSource_ID(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_vpc_attachment.test"
	resourceName := "aws_ec2_transit_gateway_vpc_attachment.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, ec2.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTransitGatewayDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayVPCAttachmentDataSourceConfig_id(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(resourceName, "appliance_mode_support", dataSourceName, "appliance_mode_support"),
					resource.TestCheckResourceAttrPair(resourceName, "dns_support", dataSourceName, "dns_support"),
					resource.TestCheckResourceAttrPair(resourceName, "ipv6_support", dataSourceName, "ipv6_support"),
					resource.TestCheckResourceAttrPair(resourceName, "subnet_ids.#", dataSourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(resourceName, "tags.%", dataSourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(resourceName, "transit_gateway_id", dataSourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_id", dataSourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(resourceName, "vpc_owner_id", dataSourceName, "vpc_owner_id"),
				),
			},
		},
	})
}

func testAccTransitGatewayVPCAttachmentDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_vpc_attachment" "test" {
  filter {
    name   = "transit-gateway-attachment-id"
    values = [aws_ec2_transit_gateway_vpc_attachment.test.id]
  }
}
`, rName))
}

func testAccTransitGatewayVPCAttachmentDataSourceConfig_id(rName string) string {
	return acctest.ConfigCompose(acctest.ConfigAvailableAZsNoOptInDefaultExclude(), fmt.Sprintf(`
resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  cidr_block        = "10.0.0.0/24"
  vpc_id            = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway" "test" {
  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_vpc_attachment" "test" {
  subnet_ids         = [aws_subnet.test.id]
  transit_gateway_id = aws_ec2_transit_gateway.test.id
  vpc_id             = aws_vpc.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_vpc_attachment" "test" {
  id = aws_ec2_transit_gateway_vpc_attachment.test.id
}
`, rName))
}
