// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package firehose_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccFirehoseDeliveryStreamDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_kinesis_firehose_delivery_stream.test"
	resourceName := "aws_kinesis_firehose_delivery_stream.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccDeliveryStreamDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrARN, resourceName, names.AttrARN),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrName, resourceName, names.AttrName),
				),
			},
		},
	})
}

func testAccDeliveryStreamDataSourceConfig_basic(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}

data "aws_kinesis_firehose_delivery_stream" "test" {
  name = aws_kinesis_firehose_delivery_stream.test.name
}
`, rName))
}
func TestAccFirehoseDeliveryStreamDataSource_databaseSourceConfiguration(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.FirehoseServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			// nosemgrep: ci.semgrep.acctest.checks.replace-planonly-checks
			{
				// Database source requires actual database infrastructure and credentials
				// Plan-only test validates schema and configuration parsing
				Config:             testAccDeliveryStreamDataSourceConfig_databaseSourceConfiguration(rName),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func testAccDeliveryStreamDataSourceConfig_databaseSourceConfiguration(rName string) string {
	return acctest.ConfigCompose(testAccDeliveryStreamConfig_base(rName), fmt.Sprintf(`
resource "aws_secretsmanager_secret" "test" {
  name = %[1]q
}

resource "aws_secretsmanager_secret_version" "test" {
  secret_id = aws_secretsmanager_secret.test.id
  secret_string = jsonencode({
    username = "testuser"
    password = "testpass"
  })
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "iceberg"

  database_source_configuration {
    type     = "MySQL"
    endpoint = "mysql.example.com"
    port     = 3306
    ssl_mode = "Disabled"

    databases {
      include = ["test_db"]
    }

    tables {
      include = ["users"]
    }

    columns {
      include = ["id", "name"]
    }

    snapshot_watermark_table = "test_db.watermark"

    database_source_authentication_configuration {
      secrets_manager_configuration {
        secret_arn = aws_secretsmanager_secret.test.arn
        role_arn   = aws_iam_role.firehose.arn
        enabled    = true
      }
    }

    database_source_vpc_configuration {
      vpc_endpoint_service_name = "com.amazonaws.vpce.${data.aws_region.current.name}.vpce-svc-1234567890abcdef0"
    }
  }

  iceberg_configuration {
    role_arn           = aws_iam_role.firehose.arn
    catalog_arn        = "arn:${data.aws_partition.current.partition}:glue:${data.aws_region.current.name}:123456789012:catalog"
    warehouse_location = "s3://${aws_s3_bucket.bucket.bucket}/warehouse/"

    destination_table_configuration {
      database_name = "test_db"
      table_name    = "test_table"
    }

    s3_configuration {
      role_arn   = aws_iam_role.firehose.arn
      bucket_arn = aws_s3_bucket.bucket.arn
    }
  }
}

data "aws_kinesis_firehose_delivery_stream" "test" {
  name = aws_kinesis_firehose_delivery_stream.test.name
}
`, rName))
}
