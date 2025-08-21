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

func TestAccVPCLatticeServiceNetworkResourceAssociationsDataSource_basic(t *testing.T) {
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
				Config: testAccServiceNetworkResourceAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-1",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-2",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-dns-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-parent-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-ip-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-1",
						"associations.*.resource_configuration_arn",
						"aws_vpclattice_resource_configuration.dns-test",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-1",
						"associations.*.resource_configuration_arn",
						"aws_vpclattice_resource_configuration.parent-test",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-2",
						"associations.*.resource_configuration_arn",
						"aws_vpclattice_resource_configuration.ip-test",
						names.AttrARN),
				),
			},
		},
	})
}

func testAccServiceNetworkResourceAssociationsDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

data "aws_caller_identity" "current" {}

data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_vpc" "test" {
  assign_generated_ipv6_cidr_block = true
  cidr_block                       = "10.0.0.0/16"
}

resource "aws_subnet" "test" {
  availability_zone = data.aws_availability_zones.available.names[0]
  vpc_id            = aws_vpc.test.id
  cidr_block        = "10.0.1.0/24"
  ipv6_cidr_block   = cidrsubnet(aws_vpc.test.ipv6_cidr_block, 8, 0)
}

resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_service_network" "test-sn-1" {
  name = "%[1]s-sn-1"
}

resource "aws_vpclattice_service_network" "test-sn-2" {
  name = "%[1]s-sn-2"
}

resource "aws_vpclattice_resource_configuration" "dns-test" {
  name = "%[1]s-dns"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = "example.com"
      ip_address_type = "IPV4"
    }
  }
}

resource "aws_vpclattice_resource_configuration" "ip-test" {
  name = "%[1]s-ip"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id

  port_ranges = ["80"]
  protocol    = "TCP"

  resource_configuration_definition {
    ip_resource {
      ip_address = "10.10.10.10"
    }
  }
}

resource "aws_vpclattice_resource_configuration" "parent-test" {
  name = "%[1]s-parent"

  protocol = "TCP"

  resource_gateway_identifier = aws_vpclattice_resource_gateway.test.id
  type                        = "GROUP"
}

resource "aws_vpclattice_resource_configuration" "child-test-1" {
  name = "%[1]s-child-1"

  port_ranges = ["80"]

  resource_configuration_group_id = aws_vpclattice_resource_configuration.parent-test.id
  type                            = "CHILD"

  resource_configuration_definition {
    ip_resource {
      ip_address = "10.10.10.10"
    }
  }
}

resource "aws_vpclattice_resource_configuration" "child-test-2" {
  name = "%[1]s-child-2"

  port_ranges = ["80"]

  resource_configuration_group_id = aws_vpclattice_resource_configuration.parent-test.id
  type                            = "CHILD"

  resource_configuration_definition {
    ip_resource {
      ip_address = "172.24.0.1"
    }
  }
}

resource "aws_vpclattice_service_network_resource_association" "dns-test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.dns-test.id
  service_network_identifier        = aws_vpclattice_service_network.test-sn-1.id
}

resource "aws_vpclattice_service_network_resource_association" "ip-test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.ip-test.id
  service_network_identifier        = aws_vpclattice_service_network.test-sn-2.id
}

resource "aws_vpclattice_service_network_resource_association" "parent-test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.parent-test.id
  service_network_identifier        = aws_vpclattice_service_network.test-sn-1.id
}

data "aws_vpclattice_service_network_resource_associations" "test-sn-1" {
  service_network_identifier = aws_vpclattice_service_network.test-sn-1.id
  depends_on                 = [aws_vpclattice_service_network_resource_association.dns-test, aws_vpclattice_service_network_resource_association.parent-test]
}

data "aws_vpclattice_service_network_resource_associations" "test-sn-2" {
  service_network_identifier = aws_vpclattice_service_network.test-sn-2.id
  depends_on                 = [aws_vpclattice_service_network_resource_association.ip-test]
}

data "aws_vpclattice_service_network_resource_associations" "test-dns-resource" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.dns-test.id
  depends_on                        = [aws_vpclattice_service_network_resource_association.dns-test]
}

data "aws_vpclattice_service_network_resource_associations" "test-ip-resource" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.ip-test.id
  depends_on                        = [aws_vpclattice_service_network_resource_association.ip-test]
}

data "aws_vpclattice_service_network_resource_associations" "test-parent-resource" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.parent-test.id
  depends_on                        = [aws_vpclattice_service_network_resource_association.parent-test]
}

`, rName)
}
