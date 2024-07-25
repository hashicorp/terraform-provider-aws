// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/timestreamwrite/types"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftimestreamwrite "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
	"github.com/hashicorp/terraform-provider-aws/internal/tfresource"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteTable_basic(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"
	dbResourceName := "aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					acctest.CheckResourceAttrRegionalARN(resourceName, names.AttrARN, "timestream", fmt.Sprintf("database/%[1]s/table/%[1]s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, names.AttrDatabaseName, dbResourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtFalse),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct0),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.enforcement_in_record", ""),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.name", ""),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.type", "MEASURE"),
					resource.TestCheckResourceAttr(resourceName, names.AttrTableName, rName),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTimestreamWriteTable_magneticStoreWriteProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_magneticStoreWriteProperties(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct0),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_magneticStoreWriteProperties(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtFalse),
				),
			},
			{
				Config: testAccTableConfig_magneticStoreWriteProperties(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_magneticStoreWriteProperties_s3Config(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_magneticStoreWritePropertiesS3(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_magneticStoreWritePropertiesS3(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rNameUpdated),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_magneticStoreWriteProperties_s3KMSConfig(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_magneticStoreWritePropertiesS3KMS(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", acctest.CtTrue),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", acctest.Ct1),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", names.AttrBucket),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.kms_key_id", "aws_kms_key.test", names.AttrARN),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.encryption_option", "SSE_KMS"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTimestreamWriteTable_disappears(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	resourceName := "aws_timestreamwrite_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftimestreamwrite.ResourceTable(), resourceName),
					acctest.CheckResourceDisappears(ctx, acctest.Provider, tftimestreamwrite.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamWriteTable_retentionProperties(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_retentionProperties(rName, 30, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "30"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.memory_store_retention_period_in_hours", "120"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_retentionProperties(rName, 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.magnetic_store_retention_period_in_days", "300"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.0.memory_store_retention_period_in_hours", "7"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", acctest.Ct1),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_tags(t *testing.T) {
	ctx := acctest.Context(t)
	var table types.Table
	resourceName := "aws_timestreamwrite_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_tags1(rName, acctest.CtKey1, acctest.CtValue1),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1),
				),
			},
			{
				Config: testAccTableConfig_tags2(rName, acctest.CtKey1, acctest.CtValue1Updated, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey1, acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct2),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", acctest.CtValue1Updated),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				Config: testAccTableConfig_tags1(rName, acctest.CtKey2, acctest.CtValue2),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsKey2, acctest.CtValue2),
					resource.TestCheckResourceAttr(resourceName, acctest.CtTagsAllPercent, acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", acctest.CtValue2),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccTimestreamWriteTable_schema(t *testing.T) {
	ctx := acctest.Context(t)
	var table1, table2 types.Table
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckTableDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccTableConfig_schema(rName, "OPTIONAL"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table1),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.enforcement_in_record", "OPTIONAL"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.name", "attr1"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.type", "DIMENSION"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableConfig_schema(rName, "REQUIRED"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(ctx, resourceName, &table2),
					testAccCheckTableNotRecreated(&table2, &table1),
					resource.TestCheckResourceAttr(resourceName, "schema.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.#", acctest.Ct1),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.enforcement_in_record", "REQUIRED"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.name", "attr1"),
					resource.TestCheckResourceAttr(resourceName, "schema.0.composite_partition_key.0.type", "DIMENSION"),
				),
			},
		},
	})
}

func testAccCheckTableDestroy(ctx context.Context) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

		for _, rs := range s.RootModule().Resources {
			if rs.Type != "aws_timestreamwrite_table" {
				continue
			}

			tableName, databaseName, err := tftimestreamwrite.TableParseResourceID(rs.Primary.ID)

			if err != nil {
				return err
			}

			_, err = tftimestreamwrite.FindTableByTwoPartKey(ctx, conn, databaseName, tableName)

			if tfresource.NotFound(err) {
				continue
			}

			if err != nil {
				return err
			}

			return fmt.Errorf("Timestream Table %s still exists", rs.Primary.ID)
		}

		return nil
	}
}

func testAccCheckTableExists(ctx context.Context, n string, v *types.Table) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		tableName, databaseName, err := tftimestreamwrite.TableParseResourceID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteClient(ctx)

		output, err := tftimestreamwrite.FindTableByTwoPartKey(ctx, conn, databaseName, tableName)

		if err != nil {
			return err
		}

		*v = *output

		return err
	}
}

func testAccCheckTableNotRecreated(i, j *types.Table) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if !aws.ToTime(i.CreationTime).Equal(aws.ToTime(j.CreationTime)) {
			return errors.New("Timestream Table was recreated")
		}

		return nil
	}
}

func testAccTableConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q
}
`, rName)
}

func testAccTableConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q
}
`, rName))
}

func testAccTableConfig_magneticStoreWriteProperties(rName string, enable bool) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = %[2]t
  }
}
`, rName, enable))
}

func testAccTableConfig_magneticStoreWritePropertiesS3(rName, prefix string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
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
`, rName, prefix))
}

func testAccTableConfig_magneticStoreWritePropertiesS3KMS(rName string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
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
`, rName))
}

func testAccTableConfig_retentionProperties(rName string, magneticStoreDays, memoryStoreHours int) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  retention_properties {
    magnetic_store_retention_period_in_days = %[2]d
    memory_store_retention_period_in_hours  = %[3]d
  }
}
`, rName, magneticStoreDays, memoryStoreHours))
}

func testAccTableConfig_tags1(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccTableConfig_tags2(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}
`, rName, tagKey1, tagValue1, tagKey2, tagValue2))
}

func testAccTableConfig_schema(rName, enforcementInRecord string) string {
	return acctest.ConfigCompose(testAccTableConfig_base(rName), fmt.Sprintf(`
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
`, rName, enforcementInRecord))
}
