// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/rds"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccRDSClusterSnapshotDataSource_dbClusterSnapshotIdentifier(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterSnapshotIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
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
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_clusterIdentifier(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "allocated_storage", resourceName, "allocated_storage"),
					resource.TestCheckResourceAttrPair(dataSourceName, "availability_zones.#", resourceName, "availability_zones.#"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine", resourceName, "engine"),
					resource.TestCheckResourceAttrPair(dataSourceName, "engine_version", resourceName, "engine_version"),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", resourceName, "kms_key_id"),
					resource.TestCheckResourceAttrPair(dataSourceName, "license_model", resourceName, "license_model"),
					resource.TestCheckResourceAttrPair(dataSourceName, "port", resourceName, "port"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshot_create_time"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshot_type", resourceName, "snapshot_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "source_db_cluster_snapshot_arn", resourceName, "source_db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "status", resourceName, "status"),
					resource.TestCheckResourceAttrPair(dataSourceName, "storage_encrypted", resourceName, "storage_encrypted"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
					resource.TestCheckResourceAttrPair(dataSourceName, "vpc_id", resourceName, "vpc_id"),
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
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
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

func TestAccRDSClusterSnapshotDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_db_cluster_snapshot.test"
	resourceName := "aws_db_cluster_snapshot.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, rds.EndpointsID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterSnapshotDataSourceConfig_tags(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_identifier", resourceName, "db_cluster_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_arn", resourceName, "db_cluster_snapshot_arn"),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_cluster_snapshot_identifier", resourceName, "db_cluster_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "tags.%", resourceName, "tags.%"),
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

func testAccClusterSnapshotDataSourceConfig_tags(rName string) string {
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
