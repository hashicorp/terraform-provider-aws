package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayConnectPeerDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect_peer.test"
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bgp_asn", resourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "inside_cidr_blocks.#", resourceName, "inside_cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_address", resourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_address", resourceName, "transit_gateway_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_attachment_id", resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_connect_peer_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeerDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_connect_peer.test"
	resourceName := "aws_ec2_transit_gateway_connect_peer.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGatewayConnect(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTransitGatewayDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayConnectPeerDataSourceConfig_id(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "bgp_asn", resourceName, "bgp_asn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "inside_cidr_blocks.#", resourceName, "inside_cidr_blocks.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_address", resourceName, "peer_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_address", resourceName, "transit_gateway_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_attachment_id", resourceName, "transit_gateway_attachment_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_connect_peer_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccTransitGatewayConnectPeerDataSourceConfig_filter(rName string) string {
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
  transit_gateway_cidr_blocks = ["10.20.30.0/24"]

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

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_connect_peer" "test" {
  filter {
    name   = "transit-gateway-connect-peer-id"
    values = [aws_ec2_transit_gateway_connect_peer.test.id]
  }
}
`, rName))
}

func testAccTransitGatewayConnectPeerDataSourceConfig_id(rName string) string {
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
  transit_gateway_cidr_blocks = ["10.20.30.0/24"]

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

resource "aws_ec2_transit_gateway_connect" "test" {
  transit_gateway_id      = aws_ec2_transit_gateway.test.id
  transport_attachment_id = aws_ec2_transit_gateway_vpc_attachment.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_connect_peer" "test" {
  inside_cidr_blocks            = ["169.254.200.0/29"]
  peer_address                  = "1.1.1.1"
  transit_gateway_attachment_id = aws_ec2_transit_gateway_connect.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_connect_peer" "test" {
  transit_gateway_connect_peer_id = aws_ec2_transit_gateway_connect_peer.test.id
}
`, rName))
}
