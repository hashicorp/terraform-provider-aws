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

	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_database.test"
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
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrCreatedTime, resourceName, names.AttrCreatedTime),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, resourceName, names.AttrKMSKeyID),
					resource.TestCheckResourceAttrPair(dataSourceName, "table_count", resourceName, "table_count"),
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
	resourceName := "aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_kmsKey(rDatabaseName, rKmsKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrDatabaseName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrARN),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabaseDataSource_updateKMSKey(t *testing.T) {
	ctx := acctest.Context(t)
	rDatabaseName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rKmsKeyName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_timestreamwrite_database.test"
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
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_kmsKey(rDatabaseName, rKmsKeyName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrKMSKeyID, kmsResourceName, names.AttrARN),
				),
			},
			{
				Config: testAccDatabaseDataSourceConfig_basic(rDatabaseName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckDatabaseExists(ctx, resourceName),
					acctest.MatchResourceAttrRegionalARN(dataSourceName, names.AttrKMSKeyID, "kms", regexache.MustCompile(`key/.+`)),
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
  name = aws_timestreamwrite_database.test.database_name
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
  name = aws_timestreamwrite_database.test.database_name
}
`, rKmsKeyName, rDatabaseName)
}
