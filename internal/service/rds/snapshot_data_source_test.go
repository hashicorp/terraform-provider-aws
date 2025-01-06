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

func TestAccRDSSnapshotDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_db_snapshot.test"
	ds1Name := "data.aws_db_snapshot.by_id"
	ds2Name := "data.aws_db_snapshot.by_tags"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.RDSServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccSnapshotDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(ds1Name, "db_instance_identifier", resourceName, "db_instance_identifier"),
					resource.TestCheckResourceAttrPair(ds1Name, "db_snapshot_arn", resourceName, "db_snapshot_arn"),
					resource.TestCheckResourceAttrPair(ds1Name, "db_snapshot_identifier", resourceName, "db_snapshot_identifier"),
					resource.TestCheckResourceAttrPair(ds1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttrSet(ds1Name, "snapshot_create_time"),
					resource.TestCheckResourceAttrSet(ds1Name, "original_snapshot_create_time"),
					resource.TestCheckResourceAttr(ds1Name, "tags.Name", rName),

					resource.TestCheckResourceAttrPair(ds2Name, "db_instance_identifier", resourceName, "db_instance_identifier"),
					resource.TestCheckResourceAttrPair(ds2Name, "db_snapshot_arn", resourceName, "db_snapshot_arn"),
					resource.TestCheckResourceAttrPair(ds2Name, "db_snapshot_identifier", resourceName, "db_snapshot_identifier"),
					resource.TestCheckResourceAttrSet(ds2Name, "snapshot_create_time"),
					resource.TestCheckResourceAttrSet(ds2Name, "original_snapshot_create_time"),
					resource.TestCheckResourceAttrPair(ds1Name, acctest.CtTagsPercent, resourceName, acctest.CtTagsPercent),
					resource.TestCheckResourceAttr(ds2Name, "tags.Name", rName),
				),
			},
		},
	})
}

func testAccSnapshotDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccSnapshotConfig_base(rName), fmt.Sprintf(`
resource "aws_db_snapshot" "test" {
  db_instance_identifier = aws_db_instance.test.identifier
  db_snapshot_identifier = %[1]q

  tags = {
    Name = %[1]q
    Test = "true"
  }
}

resource "aws_db_snapshot_copy" "test" {
  source_db_snapshot_identifier = aws_db_snapshot.test.db_snapshot_arn
  target_db_snapshot_identifier = "%[1]s-copy"

  tags = {
    Name = "%[1]s-copy"
    Test = "true"
  }
}

data "aws_db_snapshot" "by_id" {
  most_recent            = "true"
  db_snapshot_identifier = aws_db_snapshot.test.id

  depends_on = [aws_db_snapshot_copy.test]
}

data "aws_db_snapshot" "by_tags" {
  most_recent = "true"

  tags = {
    Name = %[1]q
  }

  depends_on = [aws_db_snapshot.test, aws_db_snapshot_copy.test]
}

`, rName))
}
