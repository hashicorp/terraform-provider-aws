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

func TestAccVPCNATGatewayDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameByID := "data.aws_nat_gateway.test_by_id"
	dataSourceNameBySubnetID := "data.aws_nat_gateway.test_by_subnet_id"
	dataSourceNameByTags := "data.aws_nat_gateway.test_by_tags"
	resourceName := "aws_nat_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "allocation_id", resourceName, "allocation_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_mode", resourceName, "availability_mode"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.#", resourceName, "availability_zone_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "connectivity_type", resourceName, "connectivity_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "secondary_allocation_ids.#", resourceName, "secondary_allocation_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "secondary_private_ip_address_count", resourceName, "secondary_private_ip_address_count"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "secondary_private_ip_addresses.#", resourceName, "secondary_private_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "tags.#", resourceName, "tags.#"),

					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "allocation_id", resourceName, "allocation_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "availability_mode", resourceName, "availability_mode"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "availability_zone_address.#", resourceName, "availability_zone_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "connectivity_type", resourceName, "connectivity_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "secondary_allocation_ids.#", resourceName, "secondary_allocation_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "secondary_private_ip_address_count", resourceName, "secondary_private_ip_address_count"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "secondary_private_ip_addresses.#", resourceName, "secondary_private_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameBySubnetID, "tags.#", resourceName, "tags.#"),

					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "allocation_id", resourceName, "allocation_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "availability_mode", resourceName, "availability_mode"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "availability_zone_address.#", resourceName, "availability_zone_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "connectivity_type", resourceName, "connectivity_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, names.AttrNetworkInterfaceID, resourceName, names.AttrNetworkInterfaceID),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "private_ip", resourceName, "private_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "public_ip", resourceName, "public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "secondary_allocation_ids.#", resourceName, "secondary_allocation_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "secondary_private_ip_address_count", resourceName, "secondary_private_ip_address_count"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "secondary_private_ip_addresses.#", resourceName, "secondary_private_ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByTags, "tags.#", resourceName, "tags.#"),
				),
			},
		},
	})
}

func TestAccVPCNATGatewayDataSource_availabilityModeRegional(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceNameByID := "data.aws_nat_gateway.test_by_id"
	resourceName := "aws_nat_gateway.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.EC2ServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccVPCNATGatewayDataSourceConfig_availabilityModeRegional(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "auto_provision_zones", resourceName, "auto_provision_zones"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "auto_scaling_ips", resourceName, "auto_scaling_ips"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_mode", resourceName, "availability_mode"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.#", resourceName, "availability_zone_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.0.allocation_ids.#", resourceName, "availability_zone_address.0.allocation_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.0.allocation_ids.0", resourceName, "availability_zone_address.0.allocation_ids.0"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.0.availability_zone", resourceName, "availability_zone_address.0.availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "availability_zone_address.0.availability_zone_id", resourceName, "availability_zone_address.0.availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "connectivity_type", resourceName, "connectivity_type"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.#", resourceName, "regional_nat_gateway_address.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.allocation_id", resourceName, "regional_nat_gateway_address.0.allocation_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.association_id", resourceName, "regional_nat_gateway_address.0.association_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.availability_zone", resourceName, "regional_nat_gateway_address.0.availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.availability_zone_id", resourceName, "regional_nat_gateway_address.0.availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.network_interface_id", resourceName, "regional_nat_gateway_address.0.network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.public_ip", resourceName, "regional_nat_gateway_address.0.public_ip"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "regional_nat_gateway_address.0.status", resourceName, "regional_nat_gateway_address.0.status"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "route_table_id", resourceName, "route_table_id"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, "tags.#", resourceName, "tags.#"),
					resource.TestCheckResourceAttrPair(dataSourceNameByID, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func testAccVPCNATGatewayDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccNATGatewayConfig_base(rName), fmt.Sprintf(`
resource "aws_nat_gateway" "test" {
  subnet_id     = aws_subnet.public.id
  allocation_id = aws_eip.test.id

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_internet_gateway.test]
}

data "aws_nat_gateway" "test_by_id" {
  id = aws_nat_gateway.test.id
}

data "aws_nat_gateway" "test_by_subnet_id" {
  subnet_id = aws_nat_gateway.test.subnet_id
}

data "aws_nat_gateway" "test_by_tags" {
  tags = {
    Name = aws_nat_gateway.test.tags["Name"]
  }
}
`, rName))
}

func testAccVPCNATGatewayDataSourceConfig_availabilityModeRegional(rName string) string {
	return acctest.ConfigCompose(
		testAccVPCNATGatewayConfig_availabilityModeRegionalBase(rName, 1), `
resource "aws_nat_gateway" "test" {
  vpc_id            = aws_vpc.test.id
  availability_mode = "regional"

  availability_zone_address {
    allocation_ids    = [aws_eip.test[0].id]
    availability_zone = data.aws_availability_zones.available.names[0]
  }

  depends_on = [aws_internet_gateway.test]
}

data "aws_nat_gateway" "test_by_id" {
  id = aws_nat_gateway.test.id
}
`)
}
