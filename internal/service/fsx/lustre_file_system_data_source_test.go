// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package fsx_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFSxLustreFileSystemDataSource_basic(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	ctx := acctest.Context(t)
	resourceName := "aws_fsx_lustre_file_system.test"
	datasourceName := "data.aws_fsx_lustre_file_system.test"
	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); acctest.PreCheckPartitionHasService(t, names.FSxEndpointID) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FSxServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckLustreFileSystemDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccLustreFileSystemDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(datasourceName, "automatic_backup_retention_days", resourceName, "automatic_backup_retention_days"),
					resource.TestCheckResourceAttrPair(datasourceName, "copy_tags_to_backups", resourceName, "copy_tags_to_backups"),
					resource.TestCheckResourceAttrPair(datasourceName, "data_compression_type", resourceName, "data_compression_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "deployment_type", resourceName, "deployment_type"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrDNSName, resourceName, names.AttrDNSName),
					resource.TestCheckResourceAttrPair(datasourceName, "efa_enabled", resourceName, "efa_enabled"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(datasourceName, "log_configuration.#", resourceName, "log_configuration.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "mount_name", resourceName, "mount_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "network_interface_ids.#", resourceName, "network_interface_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrOwnerID, resourceName, names.AttrOwnerID),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_capacity", resourceName, "storage_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrStorageType, resourceName, names.AttrStorageType),
					resource.TestCheckResourceAttrPair(datasourceName, "subnet_ids.#", resourceName, "subnet_ids.#"),
					resource.TestCheckResourceAttrPair(datasourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(datasourceName, "throughput_capacity", resourceName, "throughput_capacity"),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
					resource.TestCheckResourceAttrPair(datasourceName, "weekly_maintenance_start_time", resourceName, "weekly_maintenance_start_time"),
				),
			},
		},
	})
}

func testAccLustreFileSystemDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccLustreFileSystemConfig_base(rName), `
resource "aws_fsx_lustre_file_system" "test" {
  storage_capacity = 1200
  subnet_ids       = aws_subnet.test[*].id
  deployment_type  = data.aws_partition.current.partition == "aws-us-gov" ? "SCRATCH_2" : null # GovCloud does not support SCRATCH_1
}

data "aws_fsx_lustre_file_system" "test" {
  id = aws_fsx_lustre_file_system.test.id
}
`)
}
