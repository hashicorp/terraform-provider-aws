package timestreamwrite_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/timestreamwrite"
	"github.com/hashicorp/aws-sdk-go-base/v2/awsv1shim/v2/tfawserr"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/internal/conns"
	tftimestreamwrite "github.com/hashicorp/terraform-provider-aws/internal/service/timestreamwrite"
)

func TestAccTimestreamWriteTable_basic(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"
	dbResourceName := "aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					acctest.CheckResourceAttrRegionalARN(resourceName, "arn", "timestream", fmt.Sprintf("database/%[1]s/table/%[1]s", rName)),
					resource.TestCheckResourceAttrPair(resourceName, "database_name", dbResourceName, "database_name"),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "false"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "0"),
					resource.TestCheckResourceAttr(resourceName, "table_name", rName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "0"),
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
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableMagneticStoreWritePropertiesConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "0"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableMagneticStoreWritePropertiesConfig(rName, false),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "false"),
				),
			},
			{
				Config: testAccTableMagneticStoreWritePropertiesConfig(rName, true),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_magneticStoreWriteProperties_s3Config(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rNameUpdated := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableMagneticStoreWritePropertiesS3Config(rName, rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rName),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccTableMagneticStoreWritePropertiesS3Config(rName, rNameUpdated),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.object_key_prefix", rNameUpdated),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_magneticStoreWriteProperties_s3KmsConfig(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableMagneticStoreWritePropertiesS3KmsConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.enable_magnetic_store_writes", "true"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.#", "1"),
					resource.TestCheckResourceAttr(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.#", "1"),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.bucket_name", "aws_s3_bucket.test", "bucket"),
					resource.TestCheckResourceAttrPair(resourceName, "magnetic_store_write_properties.0.magnetic_store_rejected_data_location.0.s3_configuration.0.kms_key_id", "aws_kms_key.test", "arn"),
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
	resourceName := "aws_timestreamwrite_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tftimestreamwrite.ResourceTable(), resourceName),
					acctest.CheckResourceDisappears(acctest.Provider, tftimestreamwrite.ResourceTable(), resourceName),
				),
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func TestAccTimestreamWriteTable_retentionProperties(t *testing.T) {
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_table.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableRetentionPropertiesConfig(rName, 30, 120),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
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
				Config: testAccTableRetentionPropertiesConfig(rName, 300, 7),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
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
				Config: testAccTableBasicConfig(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "retention_properties.#", "1"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteTable_tags(t *testing.T) {
	resourceName := "aws_timestreamwrite_table.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); testAccPreCheck(t) },
		ErrorCheck:        acctest.ErrorCheck(t, timestreamwrite.EndpointsID),
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTableDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccTableTags1Config(rName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccTableTags2Config(rName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccTableTags1Config(rName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckTableExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(resourceName, "tags_all.key2", "value2"),
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

func testAccCheckTableDestroy(s *terraform.State) error {
	conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteConn
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "aws_timestreamwrite_table" {
			continue
		}

		tableName, dbName, err := tftimestreamwrite.TableParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		input := &timestreamwrite.DescribeTableInput{
			DatabaseName: aws.String(dbName),
			TableName:    aws.String(tableName),
		}

		output, err := conn.DescribeTableWithContext(ctx, input)

		if tfawserr.ErrCodeEquals(err, timestreamwrite.ErrCodeResourceNotFoundException) {
			continue
		}

		if err != nil {
			return err
		}

		if output != nil && output.Table != nil {
			return fmt.Errorf("Timestream Table (%s) still exists", rs.Primary.ID)
		}
	}

	return nil
}

func testAccCheckTableExists(n string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("no resource ID is set")
		}

		tableName, dbName, err := tftimestreamwrite.TableParseID(rs.Primary.ID)

		if err != nil {
			return err
		}

		conn := acctest.Provider.Meta().(*conns.AWSClient).TimestreamWriteConn

		input := &timestreamwrite.DescribeTableInput{
			DatabaseName: aws.String(dbName),
			TableName:    aws.String(tableName),
		}

		output, err := conn.DescribeTableWithContext(context.Background(), input)

		if err != nil {
			return err
		}

		if output == nil || output.Table == nil {
			return fmt.Errorf("Timestream Table (%s) not found", rs.Primary.ID)
		}

		return nil
	}
}

func testAccTableBaseConfig(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %q
}
`, rName)
}

func testAccTableBasicConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %q
}
`, rName))
}

func testAccTableMagneticStoreWritePropertiesConfig(rName string, enable bool) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %q

  magnetic_store_write_properties {
    enable_magnetic_store_writes = %t
  }
}
`, rName, enable))
}

func testAccTableMagneticStoreWritePropertiesS3Config(rName, prefix string) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
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

func testAccTableMagneticStoreWritePropertiesS3KmsConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
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

func testAccTableRetentionPropertiesConfig(rName string, magneticStoreDays, memoryStoreHours int) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %q

  retention_properties {
    magnetic_store_retention_period_in_days = %d
    memory_store_retention_period_in_hours  = %d
  }
}
`, rName, magneticStoreDays, memoryStoreHours))
}

func testAccTableTags1Config(rName, tagKey1, tagValue1 string) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
resource "aws_timestreamwrite_table" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
  table_name    = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}
`, rName, tagKey1, tagValue1))
}

func testAccTableTags2Config(rName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return acctest.ConfigCompose(
		testAccTableBaseConfig(rName),
		fmt.Sprintf(`
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
