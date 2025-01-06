// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package redshiftserverless_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRedshiftServerlessWorkgroupDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	dataSourceName := "data.aws_redshiftserverless_workgroup.test"
	resourceName := "aws_redshiftserverless_workgroup.test"

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RedshiftServerlessServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccWorkgroupDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.#", resourceName, "endpoint.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.address", resourceName, "endpoint.0.address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.port", resourceName, "endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.port", resourceName, "endpoint.0.port"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.#", resourceName, "endpoint.0.vpc_endpoint.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.vpc_endpoint_id", resourceName, "endpoint.0.vpc_endpoint.0.vpc_endpoint_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.vpc_id", resourceName, "endpoint.0.vpc_endpoint.0.vpc_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.network_interface.#", resourceName, "endpoint.0.vpc_endpoint.0.network_interface.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.network_interface.availability_zone", resourceName, "endpoint.0.vpc_endpoint.0.network_interface.availability_zone"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.network_interface.network_interface_id", resourceName, "endpoint.0.vpc_endpoint.0.network_interface.network_interface_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.network_interface.private_ip_address", resourceName, "endpoint.0.vpc_endpoint.0.network_interface.private_ip_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "endpoint.0.vpc_endpoint.0.network_interface.subnet_id", resourceName, "endpoint.0.vpc_endpoint.0.network_interface.subnet_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "enhanced_vpc_routing", resourceName, "enhanced_vpc_routing"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "namespace_name", resourceName, "namespace_name"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPubliclyAccessible, resourceName, names.AttrPubliclyAccessible),
					resource.TestCheckResourceAttrPair(dataSourceName, "security_group_ids.#", resourceName, "security_group_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workgroup_id", resourceName, "workgroup_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "workgroup_name", resourceName, "workgroup_name"),
				),
			},
		},
	})
}

func testAccWorkgroupDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_redshiftserverless_namespace" "test" {
  namespace_name = %[1]q
}

resource "aws_redshiftserverless_workgroup" "test" {
  namespace_name = aws_redshiftserverless_namespace.test.namespace_name
  workgroup_name = %[1]q
}

data "aws_redshiftserverless_workgroup" "test" {
  workgroup_name = aws_redshiftserverless_workgroup.test.workgroup_name
}
`, rName)
}
