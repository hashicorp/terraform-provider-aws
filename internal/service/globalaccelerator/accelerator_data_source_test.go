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

func TestAccGlobalAcceleratorAcceleratorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_globalaccelerator_accelerator.test"
	dataSource1Name := "data.aws_globalaccelerator_accelerator.test_by_arn"
	dataSource2Name := "data.aws_globalaccelerator_accelerator.test_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlobalAcceleratorServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSource1Name, "dual_stack_dns_name", resourceName, "dual_stack_dns_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSource1Name, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSource1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),

					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(dataSource2Name, "dual_stack_dns_name", resourceName, "dual_stack_dns_name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrEnabled, resourceName, names.AttrEnabled),
					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrHostedZoneID, resourceName, names.AttrHostedZoneID),
					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrIPAddressType, resourceName, names.AttrIPAddressType),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSource2Name, names.AttrName, resourceName, names.AttrName),
					resource.TestCheckResourceAttrPair(dataSource2Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
				),
			},
		},
	})
}

func testAccAcceleratorDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_globalaccelerator_accelerator" "test" {
  name = %[1]q

  attributes {
    flow_logs_enabled   = false
    flow_logs_s3_bucket = ""
    flow_logs_s3_prefix = "flow-logs/globalaccelerator/"
  }

  tags = {
    Name = %[1]q
  }
}

data "aws_globalaccelerator_accelerator" "test_by_arn" {
  arn = aws_globalaccelerator_accelerator.test.id
}

data "aws_globalaccelerator_accelerator" "test_by_name" {
  name = aws_globalaccelerator_accelerator.test.name
}
`, rName)
}
