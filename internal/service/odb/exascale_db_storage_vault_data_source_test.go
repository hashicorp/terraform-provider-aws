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

func TestAccODBExascaleDBStorageVaultDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var exascaleDBStorageVault odbtypes.ExascaleDbStorageVault
	rName := acctest.RandomWithPrefix(t, testAccExascaleDBStorageVaultDisplayNamePrefix)
	resourceName := "aws_odb_exascale_db_storage_vault.test"
	dataSourceName := "data.aws_odb_exascale_db_storage_vault.test"
	availabilityZoneID := testAccExascaleDBStorageVaultAvailabilityZoneIDs[acctest.Region()]

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			acctest.PreCheckRegion(t, endpoints.UsEast1RegionID, endpoints.EuWest1RegionID)
			testAccPreCheck(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.ODBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckExascaleDBStorageVaultDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccExascaleDBStorageVaultDataSourceConfig_basic(rName, availabilityZoneID),
				Check: resource.ComposeAggregateTestCheckFunc(
					testAccCheckExascaleDBStorageVaultExists(ctx, t, resourceName, &exascaleDBStorageVault),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrID, resourceName, names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zone_id", resourceName, "availability_zone_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrDisplayName, resourceName, names.AttrDisplayName),
					resource.TestCheckResourceAttr(dataSourceName, "high_capacity_database_storage.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "high_capacity_database_storage.0.total_size_in_gbs", resourceName, "high_capacity_database_storage_total_size_in_gbs"),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrStatus, "AVAILABLE"),
					resource.TestCheckResourceAttr(dataSourceName, "vm_cluster_count", "0"),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreatedAt),
					resource.TestCheckResourceAttr(dataSourceName, acctest.CtTagsPercent, "0"),
				),
			},
		},
	})
}

func testAccExascaleDBStorageVaultDataSourceConfig_basic(rName, availabilityZoneID string) string {
	return acctest.ConfigCompose(
		testAccExascaleDBStorageVaultConfig_basic(rName, availabilityZoneID),
		`
data "aws_odb_exascale_db_storage_vault" "test" {
  id = aws_odb_exascale_db_storage_vault.test.id
}
`)
}
