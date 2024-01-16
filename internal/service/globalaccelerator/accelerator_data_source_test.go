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

func TestAccGlobalAcceleratorAcceleratorDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_globalaccelerator_accelerator.test"
	dataSource1Name := "data.aws_globalaccelerator_accelerator.test_by_arn"
	dataSource2Name := "data.aws_globalaccelerator_accelerator.test_by_name"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, globalaccelerator.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccAcceleratorDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "dual_stack_dns_name", resourceName, "dual_stack_dns_name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource1Name, "tags.%", resourceName, "tags.%"),

					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.#", resourceName, "attributes.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_enabled", resourceName, "attributes.0.flow_logs_enabled"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_s3_bucket", resourceName, "attributes.0.flow_logs_s3_bucket"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "attributes.0.flow_logs_s3_prefix", resourceName, "attributes.0.flow_logs_s3_prefix"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "dual_stack_dns_name", resourceName, "dual_stack_dns_name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "enabled", resourceName, "enabled"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "hosted_zone_id", resourceName, "hosted_zone_id"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_address_type", resourceName, "ip_address_type"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.#", resourceName, "ip_sets.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.#", resourceName, "ip_sets.0.ip_addresses.#"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.0", resourceName, "ip_sets.0.ip_addresses.0"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_addresses.1", resourceName, "ip_sets.0.ip_addresses.1"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "ip_sets.0.ip_family", resourceName, "ip_sets.0.ip_family"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "name", resourceName, "name"),
					resource.TestCheckResourceAttrPair(dataSource2Name, "tags.%", resourceName, "tags.%"),
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
