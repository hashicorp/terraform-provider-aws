// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkVPCEndpointAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_vpclattice_service_network_vpc_endpoint_associations.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkVPCEndpointAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(
						dataSourceName,
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						dataSourceName,
						"associations.*.vpc_endpoint_id",
						"aws_vpc_endpoint.test-1",
						names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(
						dataSourceName,
						"associations.*.vpc_endpoint_id",
						"aws_vpc_endpoint.test-2",
						names.AttrID),
				),
			},
		},
	})
}

func testAccServiceNetworkVPCEndpointAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpc" "test-1" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"

  tags = {
    Name = "%[1]s-test-1"
  }
}

resource "aws_vpc" "test-2" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.1.0.0/16"

  tags = {
    Name = "%[1]s-test-2"
  }
}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_subnet" "test-1" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test-1.id
  cidr_block        = "10.0.1.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test-1.ipv6_cidr_block, 8, 0)

  tags = {
    Name = "%[1]s-test-1"
  }
}

resource "aws_subnet" "test-2" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test-2.id
  cidr_block        = "10.1.1.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test-2.ipv6_cidr_block, 8, 0)

  tags = {
    Name = "%[1]s-test-2"
  }
}

resource "aws_vpc_endpoint" "test-1" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test-1.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test-1.id

  tags = {
    Name = "%[1]s-test-1"
  }
}

resource "aws_vpc_endpoint" "test-2" {
  service_network_arn = aws_vpclattice_service_network.test.arn
  subnet_ids          = [aws_subnet.test-2.id]
  vpc_endpoint_type   = "ServiceNetwork"
  vpc_id              = aws_vpc.test-2.id

  tags = {
    Name = "%[1]s-test-2"
  }
}

data "aws_vpclattice_service_network_vpc_endpoint_associations" "test" {
  service_network_identifier = aws_vpclattice_service_network.test.id
  depends_on                 = [aws_vpc_endpoint.test-1, aws_vpc_endpoint.test-2]
}
`, rName)
}
