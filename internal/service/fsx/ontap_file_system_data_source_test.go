// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/service/fsx"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccFSxONTAPFileSystemDataSource_Id(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_fsx_ontap_file_system.test"
	datasourceName := "data.aws_fsx_ontap_file_system.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, fsx.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccONTAPFileSystemDataSourceConfig_id(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "automatic_backup_retention_days", resourceName, "automatic_backup_retention_days"),
					resource.TestCheckResourceAttrPair(datasourceName, "daily_automatic_backup_start_time", resourceName, "daily_automatic_backup_start_time"),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_type", resourceName, "deployment_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "disk_iops_configuration.#", resourceName, "disk_iops_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "dns_name", resourceName, "dns_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoint_ip_address_range", resourceName, "endpoint_ip_address_range"),
					resource.TestCheckResourceAttrPair(datasourceName, "endpoints.#", resourceName, "endpoints.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "owner_id", resourceName, "owner_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "preferred_subnet_id", resourceName, "preferred_subnet_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "route_table_ids.#", resourceName, "route_table_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_capacity", resourceName, "storage_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_type", resourceName, "storage_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(datasourceName, "throughput_capacity", resourceName, "throughput_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, "vpc_id", resourceName, "vpc_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "weekly_maintenance_start_time", resourceName, "weekly_maintenance_start_time"),
				),
			},
		},
	})
}

func testAccONTAPFileSystemDataSourceConfig_id(rName string) string {
	return acctest.ConfigCompose(testAccONTAPFileSystemConfig_basic(rName), `
data "aws_fsx_ontap_file_system" "test" {
  id = aws_fsx_ontap_file_system.test.id
}
`)
}
