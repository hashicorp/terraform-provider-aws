// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSClusterSnapshotDataSource_dbClusterSnapshotIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterSnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAllocatedStorage, resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageEncrypted, resourceName, names.AttrStorageEncrypted),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotDataSource_dbClusterIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAllocatedStorage, resourceName, names.AttrAllocatedStorage),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngineVersion, resourceName, names.AttrEngineVersion),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStatus, resourceName, names.AttrStatus),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrStorageEncrypted, resourceName, names.AttrStorageEncrypted),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrVPCID, resourceName, names.AttrVPCID),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotDataSource_mostRecent(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_mostRecent(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
				),
			},
		},
	})
}

func TestAccRDSClusterSnapshotDataSource_matchTags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_matchTags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(dataSourceName, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccClusterSnapshotDataSourceConfig_clusterSnapshotIdentifier(rName string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_snapshot_identifier = aws_db_cluster_snapshot.test.id
}
`, rName))
}

func testAccClusterSnapshotDataSourceConfig_clusterIdentifier(rName string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = aws_db_cluster_snapshot.test.db_cluster_identifier
}
`, rName))
}

func testAccClusterSnapshotDataSourceConfig_mostRecent(rName string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "incorrect" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = "%[1]s-incorrect"
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_db_cluster_snapshot.incorrect.db_cluster_identifier
  db_cluster_snapshot_identifier = %[1]q
}

data "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier = aws_db_cluster_snapshot.test.db_cluster_identifier
  most_recent           = true
}
`, rName))
}

func testAccClusterSnapshotDataSourceConfig_matchTags(rName string) string {
	return acctest.ConfigCompose(testAccClusterSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_cluster_snapshot" "incorrect" {
  db_cluster_identifier          = aws_rds_cluster.test.id
  db_cluster_snapshot_identifier = "%[1]s-incorrect"

  tags = {
    Name = "%[1]s-incorrect"
    Test = "true"
  }
}

resource "aws_db_cluster_snapshot" "test" {
  db_cluster_identifier          = aws_db_cluster_snapshot.incorrect.db_cluster_identifier
  db_cluster_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
    Test = "true"
  }
}

data "aws_db_cluster_snapshot" "test" {
  tags = {
    Name = %[1]q
  }

  depends_on = [aws_db_cluster_snapshot.test]
}
`, rName))
}
