// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package rds_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccRDSSnapshotsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_snapshots.test"
	resourceName := "aws_db_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotsDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "snapshots.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshots.0.db_instance_identifier", resourceName, "db_instance_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshots.0.db_snapshot_identifier", resourceName, "db_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshots.0.db_snapshot_arn", resourceName, "db_snapshot_arn"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshots.0.snapshot_create_time"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshots.0.status"),
					resource.TestCheckResourceAttrSet(dataSourceName, "snapshots.0.engine"),
					resource.TestCheckResourceAttr(dataSourceName, "snapshots.0.tags.Name", rName),
				),
			},
		},
	})
}

func TestAccRDSSnapshotsDataSource_filter(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_rds_snapshots.test"
	resourceName := "aws_db_snapshot.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDBSnapshotDestroy(ctx, t),
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotsDataSourceConfig_filter(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "snapshots.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "snapshots.0.db_snapshot_identifier", resourceName, "db_snapshot_identifier"),
				),
			},
		},
	})
}

func testAccSnapshotsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
  }
}

data "aws_rds_snapshots" "test" {
  db_instance_identifier = aws_db_snapshot.test.db_instance_identifier

  depends_on = [aws_db_snapshot.test]
}
`, rName))
}

func testAccSnapshotsDataSourceConfig_filter(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q
}

resource "aws_db_snapshot" "wrong" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = "%[1]s-wrong"
}

data "aws_rds_snapshots" "test" {
  filter {
    name   = "db-snapshot-id"
    values = [aws_db_snapshot.test.db_snapshot_identifier]
  }

  depends_on = [aws_db_snapshot.wrong]
}
`, rName))
}
