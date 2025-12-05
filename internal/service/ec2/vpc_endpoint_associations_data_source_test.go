// Copyright IBM Corp. 2014, 2025
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

func TestAccVPCEndpointAssociationsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	resourceConfigName := "aws_vpclattice_resource_configuration.test"
	datasourceName := "data.aws_vpc_endpoint_associations.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCEndpointID, resourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "associations.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.associated_resource_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "associations.0.associated_resource_arn", resourceConfigName, names.AttrARN),
					resource.TestCheckResourceAttr(datasourceName, "associations.0.dns_entry.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.dns_entry.0.dns_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.dns_entry.0.hosted_zone_id"),
				),
			},
		},
	})
}

func TestAccVPCEndpointAssociationsDataSource_serviceNetwork(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_vpc_endpoint.test"
	resourceConfigName := "aws_vpclattice_resource_configuration.test"
	resourceServiceNetwork := "aws_vpclattice_service_network.test"
	datasourceName := "data.aws_vpc_endpoint_associations.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCEndpointAssociationsDataSourceConfig_serviceNetwork(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCEndpointID, resourceName, names.AttrID),
					resource.TestCheckResourceAttr(datasourceName, "associations.#", "1"),
					resource.TestCheckResourceAttrPair(datasourceName, "associations.0.service_network_arn", resourceServiceNetwork, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "associations.0.associated_resource_arn", resourceConfigName, names.AttrARN),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.service_network_name"),
					resource.TestCheckResourceAttr(datasourceName, "associations.0.dns_entry.#", "1"),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.dns_entry.0.dns_name"),
					resource.TestCheckResourceAttrSet(datasourceName, "associations.0.dns_entry.0.hosted_zone_id"),
				),
			},
		},
	})
}

func testAccVPCEndpointAssociationsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_resourceConfiguration(rName),
		`
data "aws_vpc_endpoint_associations" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
}
`)
}

func testAccVPCEndpointAssociationsDataSourceConfig_serviceNetwork(rName string) string {
	return acctest.ConfigCompose(testAccVPCEndpointConfig_serviceNetwork(rName),
		fmt.Sprintf(`
resource "aws_vpclattice_resource_gateway" "test" {
  name       = %[1]q
  vpc_id     = aws_vpc.test.id
  subnet_ids = [aws_subnet.test.id]
}

resource "aws_vpclattice_resource_configuration" "test" {
  name = %[1]q

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

resource "aws_vpclattice_service_network_resource_association" "test" {
  resource_configuration_identifier = aws_vpclattice_resource_configuration.test.id
  service_network_identifier        = aws_vpclattice_service_network.test.id

  tags = {
    Name = %[1]q
  }
}


data "aws_vpc_endpoint_associations" "test" {
  vpc_endpoint_id = aws_vpc_endpoint.test.id
}
`, rName))
}
