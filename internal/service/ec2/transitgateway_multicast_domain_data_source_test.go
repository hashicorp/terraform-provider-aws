package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func testAccTransitGatewayMulticastDomainDataSource_Filter(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainFilterDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_accept_shared_associations", resourceName, "auto_accept_shared_associations"),
					resource.TestCheckResourceAttrPair(dataSourceName, "igmpv2_support", resourceName, "igmpv2_support"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "sources.#", "0"),
					resource.TestCheckResourceAttrPair(dataSourceName, "static_sources_support", resourceName, "static_sources_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", resourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_multicast_domain_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainDataSource_ID(t *testing.T) {
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheckTransitGateway(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainIDDataSourceConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_accept_shared_associations", resourceName, "auto_accept_shared_associations"),
					resource.TestCheckResourceAttrPair(dataSourceName, "igmpv2_support", resourceName, "igmpv2_support"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", "2"),
					resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttr(dataSourceName, "sources.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "static_sources_support", resourceName, "static_sources_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_id", resourceName, "transit_gateway_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_multicast_domain_id", resourceName, "id"),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainFilterDataSourceConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_ec2_transit_gateway" "test" {
  multicast_support = "enable"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  tags = {
    Name = %[1]q
  }
}

data "aws_ec2_transit_gateway_multicast_domain" "test" {
  filter {
    name   = "transit-gateway-multicast-domain-id"
    values = [aws_ec2_transit_gateway_multicast_domain.test.id]
  }
}
`, rName)
}

func testAccTransitGatewayMulticastDomainIDDataSourceConfig(rName string) string {
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
  multicast_support = "enable"

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

resource "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_id = aws_ec2_transit_gateway.test.id

  static_sources_support = "enable"

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_domain_association" "test" {
  subnet_id                           = aws_subnet.test.id
  transit_gateway_attachment_id       = aws_ec2_transit_gateway_vpc_attachment.test.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.test.id
}

resource "aws_network_interface" "test3" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_group_source" "test" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test3.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}

resource "aws_network_interface" "test1" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_network_interface" "test2" {
  subnet_id = aws_subnet.test.id

  tags = {
    Name = %[1]q
  }
}

resource "aws_ec2_transit_gateway_multicast_group_member" "test1" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test1.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}

resource "aws_ec2_transit_gateway_multicast_group_member" "test2" {
  group_ip_address                    = "224.0.0.1"
  network_interface_id                = aws_network_interface.test2.id
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain_association.test.transit_gateway_multicast_domain_id
}

data "aws_ec2_transit_gateway_multicast_domain" "test" {
  transit_gateway_multicast_domain_id = aws_ec2_transit_gateway_multicast_domain.test.id

  depends_on = [
    aws_ec2_transit_gateway_multicast_group_member.test1,
    aws_ec2_transit_gateway_multicast_group_member.test2,
    aws_ec2_transit_gateway_multicast_group_source.test,
  ]
}
`, rName))
}
