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
