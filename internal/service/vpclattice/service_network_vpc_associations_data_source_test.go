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

func TestAccVPCLatticeServiceNetworkVPCAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

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
				Config: testAccServiceNetworkVPCAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_vpc_associations.test-sn",
						"associations.*.vpc_id",
						"aws_vpc.test-vpc-1",
						names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_vpc_associations.test-sn",
						"associations.*.vpc_id",
						"aws_vpc.test-vpc-2",
						names.AttrID),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_vpc_associations.test-vpc-1",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_vpc_associations.test-vpc-2",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test",
						names.AttrARN),
				),
			},
		},
	})
}

func testAccServiceNetworkVPCAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

resource "aws_vpclattice_service_network" "test" {
  name = %[1]q
}

resource "aws_vpc" "test-vpc-1" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"
}

resource "aws_vpc" "test-vpc-2" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.1.0.0/16"
}

resource "aws_vpclattice_service_network_vpc_association" "test-vpc-1" {
  service_network_identifier = aws_vpclattice_service_network.test.id
  vpc_identifier             = aws_vpc.test-vpc-1.id
}

resource "aws_vpclattice_service_network_vpc_association" "test-vpc-2" {
  service_network_identifier = aws_vpclattice_service_network.test.id
  vpc_identifier             = aws_vpc.test-vpc-2.id
}

data "aws_vpclattice_service_network_vpc_associations" "test-sn" {
  service_network_identifier = aws_vpclattice_service_network.test.id
  depends_on                 = [aws_vpclattice_service_network_vpc_association.test-vpc-1, aws_vpclattice_service_network_vpc_association.test-vpc-2]
}

data "aws_vpclattice_service_network_vpc_associations" "test-vpc-1" {
  vpc_identifier = aws_vpc.test-vpc-1.id
  depends_on     = [aws_vpclattice_service_network_vpc_association.test-vpc-1]
}

data "aws_vpclattice_service_network_vpc_associations" "test-vpc-2" {
  vpc_identifier = aws_vpc.test-vpc-2.id
  depends_on     = [aws_vpclattice_service_network_vpc_association.test-vpc-2]
}

`, rName)
}
