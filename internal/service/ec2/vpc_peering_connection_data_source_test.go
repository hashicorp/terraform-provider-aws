// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package ec2_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCPeeringConnectionDataSource_cidrBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_cidrBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCIDRBlock, requesterVpcResourceName, names.AttrCIDRBlock),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_id(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_id(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					// resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block", resourceName, "cidr_block"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCIDRBlock, requesterVpcResourceName, names.AttrCIDRBlock),
					// resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block_set.#", resourceName, "cidr_block_set.#"), // not in resource
					resource.TestCheckResourceAttr(dataSourceName, "cidr_block_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cidr_block_set.*.cidr_block", requesterVpcResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(dataSourceName, "ipv6_cidr_block_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "ipv6_cidr_block_set.*.ipv6_cidr_block", requesterVpcResourceName, "ipv6_cidr_block"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "region", resourceName, "region"), // not in resource
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", resourceName, "peer_cidr_block"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", accepterVpcResourceName, names.AttrCIDRBlock),
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block_set.#", resourceName, "peer_cidr_block_set.#"), // not in resource
					resource.TestCheckResourceAttr(dataSourceName, "peer_cidr_block_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "peer_cidr_block_set.*.cidr_block", accepterVpcResourceName, names.AttrCIDRBlock),
					resource.TestCheckResourceAttr(dataSourceName, "peer_ipv6_cidr_block_set.#", acctest.Ct1),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "peer_ipv6_cidr_block_set.*.ipv6_cidr_block", accepterVpcResourceName, "ipv6_cidr_block"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_owner_id", resourceName, "peer_owner_id"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_region", resourceName, "peer_region"), //not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_vpc_id", resourceName, "peer_vpc_id"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"), // not in resource
					// resource.TestCheckResourceAttrPair(dataSourceName, "region", resourceName, "region"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_peerCIDRBlock(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_peerCIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", accepterVpcResourceName, names.AttrCIDRBlock),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_peerVPCID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_peerID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_vpc_id", resourceName, "peer_vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_vpcID(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_vpcID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionDataSourceConfig_cidrBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "requester" {
  cidr_block = "10.250.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.251.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

# aws_vpc_peering_connection does not have cidr_block
# Defer read of aws_vpc_peering_connection data source until after resource
data "aws_vpc" "requester" {
  id = aws_vpc_peering_connection.test.vpc_id
}

data "aws_vpc_peering_connection" "test" {
  cidr_block = data.aws_vpc.requester.cidr_block
}
`, rName)
}

func testAccVPCPeeringConnectionDataSourceConfig_id(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "requester" {
  cidr_block                       = "10.1.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "accepter" {
  cidr_block                       = "10.2.0.0/16"
  assign_generated_ipv6_cidr_block = true

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_peering_connection" "test" {
  id = aws_vpc_peering_connection.test.id
}
`, rName)
}

func testAccVPCPeeringConnectionDataSourceConfig_peerCIDRBlock(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "requester" {
  cidr_block = "10.252.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.253.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

# aws_vpc_peering_connection does not have cidr_block
# Defer read of aws_vpc_peering_connection data source until after resource
data "aws_vpc" "accepter" {
  id = aws_vpc_peering_connection.test.peer_vpc_id
}

data "aws_vpc_peering_connection" "test" {
  peer_cidr_block = data.aws_vpc.accepter.cidr_block
}
`, rName)
}

func testAccVPCPeeringConnectionDataSourceConfig_peerID(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "requester" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_peering_connection" "test" {
  peer_vpc_id = aws_vpc_peering_connection.test.peer_vpc_id
}
`, rName)
}

func testAccVPCPeeringConnectionDataSourceConfig_vpcID(rName string) string {
	return fmt.Sprintf(`
resource "aws_vpc" "requester" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = %[1]q
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = %[1]q
  }
}

data "aws_vpc_peering_connection" "test" {
  vpc_id = aws_vpc_peering_connection.test.vpc_id
}
`, rName)
}
