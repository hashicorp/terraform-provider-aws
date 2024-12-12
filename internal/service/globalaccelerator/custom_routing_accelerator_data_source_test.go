// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlobalAcceleratorCustomRoutingAcceleratorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	dataSource1Name := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_arn"
	dataSourceName2 := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_name"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),

					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSourceName2, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
				),
			},
		},
	})
}

func testAccCustomRoutingAcceleratorDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_custom_routing_accelerator" "test" {
  name = %[1]q
}

data "aws_globalaccelerator_custom_routing_accelerator" "test_by_arn" {
  arn = aws_globalaccelerator_custom_routing_accelerator.test.id
}

data "aws_globalaccelerator_custom_routing_accelerator" "test_by_name" {
  name = aws_globalaccelerator_custom_routing_accelerator.test.name
}
`, rName)
}
