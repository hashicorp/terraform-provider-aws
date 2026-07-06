// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package vpclattice_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccVPCLatticeServiceNetworkResourceAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	domainName := fmt.Sprintf("%s.example.com", rName)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckPartitionHasService(t, names.VPCLatticeEndpointID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccServiceNetworkResourceAssociationsDataSourceConfig_basic(rName, domainName),
				Check: resource.ComposeAggregateTestCheckFunc(
					// By service network: associations link back to the queried service network.
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
					// By service network: each resource configuration surfaces.
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
					// Flattened scalar fields are populated and correctly paired (name + status
					// together on the same element guards against a swapped-field regression).
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.aws_vpclattice_service_network_resource_associations.test-sn-1",
						"associations.*",
						map[string]string{
							"resource_configuration_name": fmt.Sprintf("%s-dns", rName),
							names.AttrStatus:              "ACTIVE",
							"is_managed_association":      acctest.CtFalse,
						}),
					// By resource configuration: query is symmetric with the by-service-network path.
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-dns-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-1",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-ip-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-2",
						names.AttrARN),
					resource.TestCheckTypeSetElemAttrPair(
						"data.aws_vpclattice_service_network_resource_associations.test-parent-resource",
						"associations.*.service_network_arn",
						"aws_vpclattice_service_network.test-sn-1",
						names.AttrARN),
					// The headline feature: the DNS name of a dns_resource association is surfaced.
					resource.TestCheckResourceAttrSet(
						"data.aws_vpclattice_service_network_resource_associations.test-dns-resource",
						"associations.0.private_dns_entry.0.domain_name"),
					// private_dns_enabled is correctly reflected on each association.
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.aws_vpclattice_service_network_resource_associations.test-dns-resource",
						"associations.*",
						map[string]string{
							"private_dns_enabled": acctest.CtTrue,
						}),
					resource.TestCheckTypeSetElemNestedAttrs(
						"data.aws_vpclattice_service_network_resource_associations.test-ip-resource",
						"associations.*",
						map[string]string{
							"private_dns_enabled": acctest.CtFalse,
						}),
				),
			},
		},
	})
}

func testAccServiceNetworkResourceAssociationsDataSourceConfig_basic(rName, domainName string) string {
	return fmt.Sprintf(`
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

  custom_domain_name = %[2]q
  port_ranges        = ["80"]
  protocol           = "TCP"

  resource_configuration_definition {
    dns_resource {
      domain_name     = %[2]q
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

resource "aws_vpclattice_service_network_resource_association" "dns-test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.dns-test.id
  service_network_identifier        = aws_vpclattice_service_network.test-sn-1.id
  private_dns_enabled               = true
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
`, rName, domainName)
}

func TestAccVPCLatticeServiceNetworkResourceAssociationsDataSource_validation(t *testing.T) {
	ctx := acctest.Context(t)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.VPCLatticeServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccServiceNetworkResourceAssociationsDataSourceConfig_missingRequired(),
				ExpectError: regexache.MustCompile(`No attribute specified when one \(and only one\) of`),
			},
			{
				Config:      testAccServiceNetworkResourceAssociationsDataSourceConfig_tooManySet(),
				ExpectError: regexache.MustCompile(`2 attributes specified when one \(and only one\) of`),
			},
		},
	})
}

func testAccServiceNetworkResourceAssociationsDataSourceConfig_missingRequired() string {
	return `
data "aws_vpclattice_service_network_resource_associations" "test" {}
`
}

func testAccServiceNetworkResourceAssociationsDataSourceConfig_tooManySet() string {
	return `
data "aws_vpclattice_service_network_resource_associations" "test" {
  service_network_identifier        = "sn-1234567890abcdef0"
  resource_configuration_identifier = "rcfg-1234567890abcdef0"
}
`
}
