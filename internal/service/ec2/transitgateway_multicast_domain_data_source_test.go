// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	tfsync "github.com/hashicorp/terraform-provider-aws/internal/experimental/sync"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccTransitGatewayMulticastDomainDataSource_Filter(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_accept_shared_associations", resourceName, "auto_accept_shared_associations"),
					resource.TestCheckResourceAttrPair(dataSourceName, "igmpv2_support", resourceName, "igmpv2_support"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(dataSourceName, "sources.#", acctest.Ct0),
					resource.TestCheckResourceAttrPair(dataSourceName, "static_sources_support", resourceName, "static_sources_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTransitGatewayID, resourceName, names.AttrTransitGatewayID),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_multicast_domain_id", resourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainDataSource_ID(t *testing.T, semaphore tfsync.Semaphore) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_ec2_transit_gateway_multicast_domain.test"
	resourceName := "aws_ec2_transit_gateway_multicast_domain.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheckTransitGatewaySynchronize(t, semaphore)
			acctest.PreCheck(ctx, t)
			testAccPreCheckTransitGateway(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccTransitGatewayMulticastDomainDataSourceConfig_iD(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "associations.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "auto_accept_shared_associations", resourceName, "auto_accept_shared_associations"),
					resource.TestCheckResourceAttrPair(dataSourceName, "igmpv2_support", resourceName, "igmpv2_support"),
					resource.TestCheckResourceAttr(dataSourceName, "members.#", acctest.Ct2),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttr(dataSourceName, "sources.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "static_sources_support", resourceName, "static_sources_support"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrTransitGatewayID, resourceName, names.AttrTransitGatewayID),
					resource.TestCheckResourceAttrPair(dataSourceName, "transit_gateway_multicast_domain_id", resourceName, names.AttrID),
				),
			},
		},
	})
}

func testAccTransitGatewayMulticastDomainDataSourceConfig_filter(rName string) string {
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

func testAccTransitGatewayMulticastDomainDataSourceConfig_iD(rName string) string {
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
