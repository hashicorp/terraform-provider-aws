package ec2_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDataSourceAwsVpcPeeringConnection_CidrBlock(t *testing.T) {
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionCIDRBlockDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block", requesterVpcResourceName, "cidr_block"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcPeeringConnection_Id(t *testing.T) {
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionIDDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block", resourceName, "cidr_block"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block", requesterVpcResourceName, "cidr_block"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block_set.#", resourceName, "cidr_block_set.#"), // not in resource
					resource.TestCheckResourceAttr(dataSourceName, "cidr_block_set.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "cidr_block_set.*.cidr_block", requesterVpcResourceName, "cidr_block"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "region", resourceName, "region"), // not in resource
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", resourceName, "peer_cidr_block"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", accepterVpcResourceName, "cidr_block"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block_set.#", resourceName, "peer_cidr_block_set.#"), // not in resource
					resource.TestCheckResourceAttr(dataSourceName, "peer_cidr_block_set.#", "1"),
					resource.TestCheckTypeSetElemAttrPair(dataSourceName, "peer_cidr_block_set.*.cidr_block", accepterVpcResourceName, "cidr_block"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_owner_id", resourceName, "peer_owner_id"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "peer_region", resourceName, "peer_region"), //not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_vpc_id", resourceName, "peer_vpc_id"),
					// resource.TestCheckResourceAttrPair(dataSourceName, "owner_id", resourceName, "owner_id"), // not in resource
					// resource.TestCheckResourceAttrPair(dataSourceName, "region", resourceName, "region"), // not in resource
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcPeeringConnection_PeerCidrBlock(t *testing.T) {
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionPeerCIDRBlockDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", accepterVpcResourceName, "cidr_block"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcPeeringConnection_PeerVpcId(t *testing.T) {
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionPeerVpcIDDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_vpc_id", resourceName, "peer_vpc_id"),
				),
			},
		},
	})
}

func TestAccDataSourceAwsVpcPeeringConnection_VpcId(t *testing.T) {
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:   func() { acctest.PreCheck(t) },
		ErrorCheck: acctest.ErrorCheck(t, ec2.EndpointsID),
		Providers:  acctest.Providers,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionVpcIDDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
				),
			},
		},
	})
}

func testAccVPCPeeringConnectionCIDRBlockDataSourceConfig() string {
	return `
resource "aws_vpc" "requester" {
  cidr_block = "10.250.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.251.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
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
`
}

func testAccVPCPeeringConnectionIDDataSourceConfig() string {
	return `
resource "aws_vpc" "requester" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

data "aws_vpc_peering_connection" "test" {
  id = aws_vpc_peering_connection.test.id
}
`
}

func testAccVPCPeeringConnectionPeerCIDRBlockDataSourceConfig() string {
	return `
resource "aws_vpc" "requester" {
  cidr_block = "10.252.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.253.0.0/16" # CIDR must be different than other tests

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
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
`
}

func testAccVPCPeeringConnectionPeerVpcIDDataSourceConfig() string {
	return `
resource "aws_vpc" "requester" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

data "aws_vpc_peering_connection" "test" {
  peer_vpc_id = aws_vpc_peering_connection.test.peer_vpc_id
}
`
}

func testAccVPCPeeringConnectionVpcIDDataSourceConfig() string {
	return `
resource "aws_vpc" "requester" {
  cidr_block = "10.1.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc" "accepter" {
  cidr_block = "10.2.0.0/16"

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

resource "aws_vpc_peering_connection" "test" {
  vpc_id      = aws_vpc.requester.id
  peer_vpc_id = aws_vpc.accepter.id
  auto_accept = true

  tags = {
    Name = "terraform-testacc-vpc-peering-connection-data-source"
  }
}

data "aws_vpc_peering_connection" "test" {
  vpc_id = aws_vpc_peering_connection.test.vpc_id
}
`
}
