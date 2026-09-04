// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package odb_test

import (
	"testing"

	odbtypes "github.com/aws/aws-sdk-go-v2/service/odb/types"
	"github.com/hashicorp/aws-sdk-go-base/v2/endpoints"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccODBExaDBVMClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exaDBVMCluster odbtypes.ExadbVmCluster
	rName := testAccRandomExaDBVMClusterDisplayName(t)
	hostname := testAccRandomExaDBVMClusterHostname(t)
	gridImageID := testAccExaDBVMClusterGridImageIDForRegion(ctx, t, endpoints.UsEast1RegionID, testAccExaDBVMClusterAvailabilityZoneID)
	publicKey := testAccRandomExaDBVMClusterSSHPublicKey(t)
	resourceName := "aws_odb_exadb_vm_cluster.test"
	dataSourceName := "data.aws_odb_exadb_vm_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID)
			testAccPreCheckExaDBVMCluster(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExaDBVMClusterDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExaDBVMClusterDataSourceConfig_basic(rName, hostname, gridImageID, publicKey),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExaDBVMClusterExists(ctx, t, resourceName, &exaDBVMCluster),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedAt, resourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttrPair(dataSourceName, "enabled_ecpu_count", resourceName, "enabled_ecpu_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "exascale_db_storage_vault_arn", resourceName, "exascale_db_storage_vault_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "exascale_db_storage_vault_id", resourceName, "exascale_db_storage_vault_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "grid_image_id", resourceName, "grid_image_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "hostname", resourceName, "hostname"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_count", resourceName, "node_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "odb_network_arn", resourceName, "odb_network_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "odb_network_id", resourceName, "odb_network_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "shape", resourceName, "shape"),
					resource.TestCheckResourceAttr(dataSourceName, "ssh_public_keys.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "AVAILABLE"),
					resource.TestCheckResourceAttrPair(dataSourceName, "total_ecpu_count", resourceName, "total_ecpu_count"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vm_file_system_storage_total_size_in_gbs", resourceName, "vm_file_system_storage_total_size_in_gbs"),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccExaDBVMClusterDataSourceConfig_basic(rName, hostname, gridImageID, publicKey string) string {
	return acctest.ConfigCompose(
		testAccExaDBVMClusterConfig_basic(rName, hostname, gridImageID, publicKey),
		`
data "aws_odb_exadb_vm_cluster" "test" {
  id = aws_odb_exadb_vm_cluster.test.id
}
`)
}
