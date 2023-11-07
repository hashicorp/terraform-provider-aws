// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package globalaccelerator_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/globalaccelerator"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccGlobalAcceleratorCustomRoutingAcceleratorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_globalaccelerator_custom_routing_accelerator.test"
	dataSource1Name := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_arn"
	dataSourceName2 := "data.aws_globalaccelerator_custom_routing_accelerator.test_by_name"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCustomRoutingAcceleratorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),

					resource.TestCheckResourceAttrPair(dataSourceName2, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSourceName2, "hosted_zone_id", resourceName, "hosted_zone_id"),
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
