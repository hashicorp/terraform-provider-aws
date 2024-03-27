// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"fmt"
	"testing"

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
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_magneticStoreWriteProperties(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "0"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "false"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_magneticStoreWriteProperties_s3Config(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWritePropertiesS3(rDatabaseName, rTableName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rTableName),
				),
			},
			{
				Config: testAccTableDataSourceConfig_magneticStoreWritePropertiesS3(rDatabaseName, rTableName, rTableNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rTableNameUpdated),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_magneticStoreWriteProperties_s3KMSConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWritePropertiesS3KMS(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rTableName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.encryption_option", "SSE_KMS"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_retentionProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_retentionProperties(rDatabaseName, rTableName, 30, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.memory_store_retention_period_in_hours", "120"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_retentionProperties(rDatabaseName, rTableName, 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "300"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.memory_store_retention_period_in_hours", "7"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_basic(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", "1"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_tags1(rDatabaseName, rTableName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_tags2(rDatabaseName, rTableName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_tags1(rDatabaseName, rTableName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_schema(t *testing.T) {
	ctx := acctest.Context(t)
	var table1, table2 types.Table
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_schema(rDatabaseName, rTableName, "OPTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table1),
					resource.TestCheckResourceAttr(dataSourceName, "schema.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.enforcement_in_record", "OPTIONAL"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.name", "attr1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.type", "DIMENSION"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_schema(rDatabaseName, rTableName, "REQUIRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, dataSourceName, &table2),
					testAccCheckTableNotRecreated(&table2, &table1),
					resource.TestCheckResourceAttr(dataSourceName, "schema.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.#", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.enforcement_in_record", "REQUIRED"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.name", "attr1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.type", "DIMENSION"),
				),
			},
		},
	})
}

func testAccTableDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}
`, rName)
}

func testAccTableDataSourceConfig_basic(rDatabaseName, rTableName string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name    = %[2]q
  }
 
data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }

  `, rDatabaseName, rTableName))
}

func testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName string, enable bool) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name    = %[1]q

	magnetic_store_write_properties {
		enable_magnetic_store_writes = %[2]t
	  }
  }

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }

  `, rTableName, enable))
}

func testAccTableDataSourceConfig_magneticStoreWritePropertiesS3(rDatabaseName, rTableName, prefix string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true

    magnetic_store_rejected_data_location {
      s3_configuration {
        bucket_name       = aws_s3_bucket.test.bucket
        object_key_prefix = %[2]q
      }
    }
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName, prefix))
}

func testAccTableDataSourceConfig_magneticStoreWritePropertiesS3KMS(rDatabaseName, rTableName string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_s3_bucket" "test" {
  bucket        = %[1]q
  force_destroy = true
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true

    magnetic_store_rejected_data_location {
      s3_configuration {
        bucket_name       = aws_s3_bucket.test.bucket
        object_key_prefix = %[1]q
        kms_key_id        = aws_kms_key.test.arn
        encryption_option = "SSE_KMS"
      }
    }
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName))
}

func testAccTableDataSourceConfig_retentionProperties(rDatabaseName, rTableName string, magneticStoreDays, memoryStoreHours int) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  retention_properties {
    magnetic_store_retention_period_in_days = %[2]d
    memory_store_retention_period_in_hours  = %[3]d
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName, magneticStoreDays, memoryStoreHours))
}

func testAccTableDataSourceConfig_tags1(rDatabaseName, rTableName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName, tagKey1, tagValue1))
}

func testAccTableDataSourceConfig_tags2(rDatabaseName, rTableName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTableDataSourceConfig_schema(rDatabaseName, rTableName, enforcementInRecord string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  schema {
    composite_partition_key {
      enforcement_in_record = %[2]q
      name                  = "attr1"
      type                  = "DIMENSION"
    }
  }
}

data "aws_timestreamwrite_table" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
	table_name = aws_timestreamwrite_table.test.table_name
  }
`, rTableName, enforcementInRecord))
}
