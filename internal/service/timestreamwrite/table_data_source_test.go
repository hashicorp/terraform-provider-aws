// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteTableDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_basic(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "database_name", rDatabaseName),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rTableName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "false"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "0"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.enforcement_in_record", ""),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.name", ""),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.type", "MEASURE"),
					resource.TestCheckResourceAttr(dataSourceName, "table_name", rTableName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),

					acctest.MatchResourceAttrRegionalARN(dataSourceName, "arn", "timestreamwrite", regexache.MustCompile(`table:+.`)),
				),
			},
		},
	})
}

func testAccTableDataSourceConfig_basic(rName, rTableName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
	database_name = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name    = %[2]q
  }
 
 data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.id
}

  `, rName, rTableName)
}
