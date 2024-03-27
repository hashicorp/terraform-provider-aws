// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package timestreamwrite_test

import (
	"fmt"
	"testing"

	"github.com/YakDriver/regexache"
	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"

	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccTimestreamWriteDatabaseDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)

	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	// var database types.Database
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_basic(rDatabaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					acctest.CheckResourceAttrRegionalARN(dataSourceName, "arn", "timestream", fmt.Sprintf("database/%s", rDatabaseName)),
					resource.TestCheckResourceAttr(dataSourceName, "database_name", rDatabaseName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "kms_key_id", "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "0"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabaseDataSource_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rKmsKeyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_kmsKey(rDatabaseName, rKmsKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "database_name", rDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", kmsResourceName, "arn"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabaseDataSource_updateKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rKmsKeyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_basic(rDatabaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "kms_key_id", "kms", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_kmsKey(rDatabaseName, rKmsKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", kmsResourceName, "arn"),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_basic(rDatabaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, "kms_key_id", "kms", regexache.MustCompile(`key/.+`)),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabaseDataSource_tags(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_tags1(rDatabaseName, "key1", "value1"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key1", "value1"),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_tags2(rDatabaseName, "key1", "value1updated", "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key1", "value1updated"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key1", "value1updated"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key2", "value2"),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_tags1(rDatabaseName, "key2", "value2"),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, dataSourceName),
					resource.TestCheckResourceAttr(dataSourceName, "tags.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags.key2", "value2"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.%", "1"),
					resource.TestCheckResourceAttr(dataSourceName, "tags_all.key2", "value2"),
				),
			},
		},
	})
}

func testAccDatabaseDataSourceConfig_basic(rDatabaseName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
		database_name = %[1]q
  }

data "aws_timestreamwrite_database" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
  }
`, rDatabaseName)
}

func testAccDatabaseDataSourceConfig_kmsKey(rDatabaseName, rKmsKeyName string) string {
	return fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description = %[1]q

  policy = <<POLICY
{
 "Version": "2012-10-17",
 "Statement": [
   {
     "Effect": "Allow",
     "Principal": {
       "AWS": "*"
     },
     "Action": "kms:*",
     "Resource": "*"
   }
 ]
}
POLICY
}

resource "aws_timestreamwrite_database" "test" {
  database_name = %[2]q
  kms_key_id    = aws_kms_key.test.arn
}

data "aws_timestreamwrite_database" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
  }
`, rKmsKeyName, rDatabaseName)
}

func testAccDatabaseDataSourceConfig_tags1(rDatabaseName, tagKey1, tagValue1 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
  }
}

data "aws_timestreamwrite_database" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
  }
`, rDatabaseName, tagKey1, tagValue1)
}

func testAccDatabaseDataSourceConfig_tags2(rDatabaseName, tagKey1, tagValue1, tagKey2, tagValue2 string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
  database_name = %[1]q

  tags = {
    %[2]q = %[3]q
    %[4]q = %[5]q
  }
}

data "aws_timestreamwrite_database" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
  }
`, rDatabaseName, tagKey1, tagValue1, tagKey2, tagValue2)
}
