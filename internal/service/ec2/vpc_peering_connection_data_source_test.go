package ec2_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/ec2"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccVPCPeeringConnectionDataSource_cidrBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_cidrBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "cidr_block", requesterVpcResourceName, "cidr_block"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_id(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"
	requesterVpcResourceName := "aws_vpc.requester"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_id(rName),
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

func TestAccVPCPeeringConnectionDataSource_peerCIDRBlock(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"
	accepterVpcResourceName := "aws_vpc.accepter"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_peerCIDRBlock(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_cidr_block", accepterVpcResourceName, "cidr_block"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_peerVPCID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_peerID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "peer_vpc_id", resourceName, "peer_vpc_id"),
				),
			},
		},
	})
}

func TestAccVPCPeeringConnectionDataSource_vpcID(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpc_peering_connection.test"
	resourceName := "aws_vpc_peering_connection.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, ec2.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCPeeringConnectionDataSourceConfig_vpcID(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "id", resourceName, "id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
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
