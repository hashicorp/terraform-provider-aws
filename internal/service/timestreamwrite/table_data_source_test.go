// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	"github.com/hashicorp/terraform-provider-aws/internal/errs"
	tftimestreamwrite "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteTableDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_basic(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrSet(dataSourceName, names.AttrCreationTime),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrDatabaseName, rDatabaseName),
					resource.TestCheckResourceAttr(dataSourceName, names.AttrName, rTableName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtFalse),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct0),
					resource.TestCheckResourceAttrSet(dataSourceName, "retention_properties.#"),
					resource.TestCheckResourceAttrSet(dataSourceName, "schema.#"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "table_status", resourceName, names.AttrStatus),
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

	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct0),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_magneticStoreWriteProperties_s3Config(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWritePropertiesS3(rDatabaseName, rTableName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rTableName),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_magneticStoreWriteProperties_s3KMSConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_magneticStoreWritePropertiesS3KMS(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttrPair(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rTableName),
					resource.TestCheckResourceAttr(dataSourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.encryption_option", "SSE_KMS"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_retentionProperties(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_retentionProperties(rDatabaseName, rTableName, 30, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "30"),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.0.memory_store_retention_period_in_hours", "120"),
				),
			},
			{
				Config: testAccTableDataSourceConfig_basic(rDatabaseName, rTableName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "retention_properties.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTableDataSource_schema(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rTableName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDatasourceDestroy(ctx, rDatabaseName, rTableName),
		Steps: []resource.TestStep{
			{
				Config: testAccTableDataSourceConfig_schema(rDatabaseName, rTableName, "OPTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExistsNames(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.#", acctest.Ct1),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.enforcement_in_record", "OPTIONAL"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.name", "attr1"),
					resource.TestCheckResourceAttr(dataSourceName, "schema.0.composite_partition_key.0.type", "DIMENSION"),
				),
			},
		},
	})
}
func testAccCheckTableDatasourceDestroy(ctx context.Context, databaseName string, tableName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)
		_, err := tftimestreamwrite.FindTableByTwoPartKey(ctx, conn, databaseName, tableName)

		if errs.IsA[*types.ResourceNotFoundException](err) {
			return nil
		}
		if err != nil {
			return err
		}

		return fmt.Errorf(("Timestream Table %s still exists" + databaseName + " " + tableName))
	}
}
func testAccCheckTableExistsNames(ctx context.Context, n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

		_, err := tftimestreamwrite.FindTableByTwoPartKey(ctx, conn, rs.Primary.Attributes[names.AttrDatabaseName], rs.Primary.Attributes[names.AttrName])

		if err != nil {
			return err
		}
		return err
	}
}

func testAccTableDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}




`, rName)
}

func testAccTableDataSourceConfig_basic(rDatabaseName, rTableName string) string {
	return acctest.ConfigCompose(testAccTableDataSourceConfig_base(rDatabaseName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q
}

  `, rDatabaseName, rTableName))
}

func testAccTableDataSourceConfig_magneticStoreWriteProperties(rDatabaseName, rTableName string, enable bool) string {
	return fmt.Sprintf(`


resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = %[3]t
  }
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}




  `, rDatabaseName, rTableName, enable)
}

func testAccTableDataSourceConfig_magneticStoreWritePropertiesS3(rDatabaseName, rTableName, prefix string) string {
	return fmt.Sprintf(`


resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true

    magnetic_store_rejected_data_location {
      s3_configuration {
        bucket_name       = aws_s3_bucket.test.bucket
        object_key_prefix = %[3]q
      }
    }
  }
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}
`, rDatabaseName, rTableName, prefix)
}

func testAccTableDataSourceConfig_magneticStoreWritePropertiesS3KMS(rDatabaseName, rTableName string) string {
	return fmt.Sprintf(`


resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_s3_bucket" "test" {
  bucket        = %[2]q
  force_destroy = true
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  description             = %[2]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = true

    magnetic_store_rejected_data_location {
      s3_configuration {
        bucket_name       = aws_s3_bucket.test.bucket
        object_key_prefix = %[2]q
        kms_key_id        = aws_kms_key.test.arn
        encryption_option = "SSE_KMS"
      }
    }
  }
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}
`, rDatabaseName, rTableName)
}

func testAccTableDataSourceConfig_retentionProperties(rDatabaseName, rTableName string, magneticStoreDays, memoryStoreHours int) string {
	return fmt.Sprintf(`




resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q

  retention_properties {
    magnetic_store_retention_period_in_days = %[3]d
    memory_store_retention_period_in_hours  = %[4]d
  }
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}
`, rDatabaseName, rTableName, magneticStoreDays, memoryStoreHours)
}

func testAccTableDataSourceConfig_schema(rDatabaseName, rTableName, enforcementInRecord string) string {
	return fmt.Sprintf(`




resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}

resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[2]q

  schema {
    composite_partition_key {
      enforcement_in_record = %[3]q
      name                  = "attr1"
      type                  = "DIMENSION"
    }
  }
}

data "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  name          = aws_timestreamwrite_table.test.table_name
}
`, rDatabaseName, rTableName, enforcementInRecord)
}
