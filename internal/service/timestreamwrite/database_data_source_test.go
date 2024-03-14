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

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"
	resourceName := "aws_timestreamwrite_database.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             nil,
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, "arn", resourceName, "arn"),
					resource.TestCheckResourceAttr(dataSourceName, "database_name", rName),
					acctest.MatchResourceAttrRegionalARN(resourceName, "kms_key_id", "kms", regexache.MustCompile(`key/.+`)),
					resource.TestCheckResourceAttr(dataSourceName, "table_count", "0"),
				),
			},
		},
	})
}

func TestAccTimestreamWriteDatabaseDataSource_kmsKey(t *testing.T) {
	ctx := acctest.Context(t)
	resourceName := "aws_timestreamwrite_database.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_timestreamwrite_database.test"
	kmsResourceName := "aws_kms_key.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t); testAccPreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.TimestreamWriteServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		CheckDestroy:             testAccCheckDatabaseDestroy(ctx),
		Steps: []resource.TestStep{
			{
				Config: testAccDatabaseDataSourceConfig_kmsKey(rName),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, "database_name", rName),
					resource.TestCheckResourceAttrPair(dataSourceName, "kms_key_id", kmsResourceName, "arn"),
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

func testAccDatabaseDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_timestreamwrite_database" "test" {
	database_name = %[1]q
}

data "aws_timestreamwrite_database" "test" {
	database_name = aws_timestreamwrite_database.test.database_name
}
`, rName)
}

func testAccDatabaseDataSourceConfig_kmsKey(rName string) string {
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
  database_name = %[1]q
  kms_key_id    = aws_kms_key.test.arn
}

data "aws_timestreamwrite_database" "test" {
  database_name = aws_timestreamwrite_database.test.database_name
}
`, rName)
}
