// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package glue_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccGlueCatalogTablesDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_glue_catalog_tables.test"

	dbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTablesDataSourceConfig_basic(dbName, tName1, tName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrDatabaseName, dbName),
					resource.TestCheckResourceAttr(datasourceName, "tables.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "tables.0.name", tName1),
					resource.TestCheckResourceAttr(datasourceName, "tables.1.name", tName2),
					resource.TestCheckResourceAttr(datasourceName, "tables.0.table_type", "EXTERNAL_TABLE"),
					resource.TestCheckResourceAttr(datasourceName, "tables.1.table_type", "EXTERNAL_TABLE"),
					resource.TestCheckResourceAttrSet(datasourceName, "tables.0.arn"),
					resource.TestCheckResourceAttrSet(datasourceName, "tables.1.arn"),
				),
			},
		},
	})
}

func TestAccGlueCatalogTablesDataSource_expression(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_glue_catalog_tables.test"

	dbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tName1 := "table_test_1"
	tName2 := "table_test_2"
	tName3 := "other_table"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTablesDataSourceConfig_expression(dbName, tName1, tName2, tName3),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrDatabaseName, dbName),
					resource.TestCheckResourceAttr(datasourceName, "expression", "table_test_.*"),
					resource.TestCheckResourceAttr(datasourceName, "tables.#", "2"),
					resource.TestCheckResourceAttr(datasourceName, "tables.0.name", tName1),
					resource.TestCheckResourceAttr(datasourceName, "tables.1.name", tName2),
				),
			},
		},
	})
}

func TestAccGlueCatalogTablesDataSource_tableType(t *testing.T) {
	ctx := acctest.Context(t)
	datasourceName := "data.aws_glue_catalog_tables.test"

	dbName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tName1 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	tName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.GlueServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCatalogTablesDataSourceConfig_tableType(dbName, tName1, tName2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(datasourceName, names.AttrDatabaseName, dbName),
					resource.TestCheckResourceAttr(datasourceName, "table_type", "EXTERNAL_TABLE"),
					resource.TestCheckResourceAttr(datasourceName, "tables.#", "1"),
					resource.TestCheckResourceAttr(datasourceName, "tables.0.name", tName1),
					resource.TestCheckResourceAttr(datasourceName, "tables.0.table_type", "EXTERNAL_TABLE"),
				),
			},
		},
	})
}

func testAccCatalogTablesDataSourceConfig_basic(dbName, tName1, tName2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test1" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[2]q

  description = "Test table 1"
  table_type  = "EXTERNAL_TABLE"

  parameters = {
    EXTERNAL              = "TRUE"
    "parquet.compression" = "SNAPPY"
  }

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream1"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream1"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"

      parameters = {
        "serialization.format" = 1
      }
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

resource "aws_glue_catalog_table" "test2" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[3]q

  description = "Test table 2"
  table_type  = "EXTERNAL_TABLE"

  parameters = {
    EXTERNAL              = "TRUE"
    "parquet.compression" = "SNAPPY"
  }

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream2"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream2"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"

      parameters = {
        "serialization.format" = 1
      }
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

data "aws_glue_catalog_tables" "test" {
  database_name = aws_glue_catalog_database.test.name

  depends_on = [
    aws_glue_catalog_table.test1,
    aws_glue_catalog_table.test2,
  ]
}
`, dbName, tName1, tName2)
}

func testAccCatalogTablesDataSourceConfig_expression(dbName, tName1, tName2, tName3 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test1" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[2]q

  description = "Test table 1"
  table_type  = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream1"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream1"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

resource "aws_glue_catalog_table" "test2" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[3]q

  description = "Test table 2"
  table_type  = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream2"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream2"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

resource "aws_glue_catalog_table" "test3" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[4]q

  description = "Other table"
  table_type  = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream3"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream3"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

data "aws_glue_catalog_tables" "test" {
  database_name = aws_glue_catalog_database.test.name
  expression    = "table_test_.*"

  depends_on = [
    aws_glue_catalog_table.test1,
    aws_glue_catalog_table.test2,
    aws_glue_catalog_table.test3,
  ]
}
`, dbName, tName1, tName2, tName3)
}

func testAccCatalogTablesDataSourceConfig_tableType(dbName, tName1, tName2 string) string {
	return fmt.Sprintf(`
resource "aws_glue_catalog_database" "test" {
  name = %[1]q
}

resource "aws_glue_catalog_table" "test1" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[2]q

  description = "External table"
  table_type  = "EXTERNAL_TABLE"

  storage_descriptor {
    location      = "s3://my-bucket/event-streams/my-stream1"
    input_format  = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetInputFormat"
    output_format = "org.apache.hadoop.hive.ql.io.parquet.MapredParquetOutputFormat"

    ser_de_info {
      name                  = "my-stream1"
      serialization_library = "org.apache.hadoop.hive.ql.io.parquet.serde.ParquetHiveSerDe"
    }

    columns {
      name = "my_string"
      type = "string"
    }
  }
}

resource "aws_glue_catalog_table" "test2" {
  database_name = aws_glue_catalog_database.test.name
  name          = %[3]q

  description = "Virtual view"
  table_type  = "VIRTUAL_VIEW"

  view_original_text = "SELECT * FROM test1"
  view_expanded_text = "SELECT my_string FROM test1"

  storage_descriptor {
    columns {
      name = "my_string"
      type = "string"
    }
  }
}

data "aws_glue_catalog_tables" "test" {
  database_name = aws_glue_catalog_database.test.name
  table_type    = "EXTERNAL_TABLE"

  depends_on = [
    aws_glue_catalog_table.test1,
    aws_glue_catalog_table.test2,
  ]
}
`, dbName, tName1, tName2)
}
