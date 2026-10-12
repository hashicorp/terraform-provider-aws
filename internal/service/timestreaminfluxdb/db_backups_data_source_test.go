// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package timestreaminfluxdb_test

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamInfluxDBDBBackupsDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreaminfluxdb_db_backups.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(ctx, t)
			testAccPreCheckDBBackups(ctx, t)
		},
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamInfluxDBServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDBBackupsDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "backups.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "backups.0.id", "aws_timestreaminfluxdb_db_backup.test", names.AttrID),
					resource.TestCheckResourceAttrPair(dataSourceName, "backups.0.name", "aws_timestreaminfluxdb_db_backup.test", names.AttrName),
					resource.TestCheckResourceAttrPair(dataSourceName, "db_resource_id", "aws_timestreaminfluxdb_db_instance.test", names.AttrID),
				),
			},
		},
	})
}

func testAccDBBackupsDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDBBackupConfig_basic(rName), `
data "aws_timestreaminfluxdb_db_backups" "test" {
  db_resource_id = aws_timestreaminfluxdb_db_instance.test.id

  depends_on = [aws_timestreaminfluxdb_db_backup.test]
}
`)
}
