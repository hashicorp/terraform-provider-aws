package dms_test

import (
	"fmt"
	"testing"

	dms "github.com/aws/aws-sdk-go/service/databasemigrationservice"
	sdkacctest "github.com/hashicorp/terraform-plugin-sdk/v2/helper/acctest"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
)

func TestAccDMSS3Endpoint_basic(t *testing.T) {
	resourceName := "aws_dms_s3_endpoint.test"
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:     func() { acctest.PreCheck(t) },
		ErrorCheck:   acctest.ErrorCheck(t, dms.EndpointsID),
		Providers:    acctest.Providers,
		CheckDestroy: testAccCheckEndpointDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccS3EndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", "bucket_name"),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", "true"),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "100"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", "x"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "?"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\r\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", "UNDERSCORE"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", "yyyymmddhh"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", "Asia/Seoul"),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", "false"),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttr(resourceName, "external_table_definition", "etd"),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", "1"),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", "true"),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", "parquet-2-0"),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", "false"),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", "false"),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "11000"),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.iam_role", "arn"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time"),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", "true"),
				),
			},
			{
				ResourceName:      resourceName,
				ImportState:       true,
				ImportStateVerify: true,
			},
			{
				Config: testAccS3EndpointConfig_basic(rName),
				Check: resource.ComposeTestCheckFunc(
					testAccCheckEndpointExists(resourceName),
					resource.TestCheckResourceAttr(resourceName, "add_column_name", "true"),
					resource.TestCheckResourceAttr(resourceName, "bucket_folder", "folder"),
					resource.TestCheckResourceAttr(resourceName, "bucket_name", "updated_name"),
					resource.TestCheckResourceAttr(resourceName, "canned_acl_for_objects", "private"),
					resource.TestCheckResourceAttr(resourceName, "cdc_inserts_and_updates", "true"),
					resource.TestCheckResourceAttr(resourceName, "cdc_max_batch_interval", "100"),
					resource.TestCheckResourceAttr(resourceName, "cdc_min_file_size", "16"),
					resource.TestCheckResourceAttr(resourceName, "cdc_path", "cdc/path"),
					resource.TestCheckResourceAttr(resourceName, "compression_type", "GZIP"),
					resource.TestCheckResourceAttr(resourceName, "csv_delimiter", ";"),
					resource.TestCheckResourceAttr(resourceName, "csv_no_sup_value", "x"),
					resource.TestCheckResourceAttr(resourceName, "csv_null_value", "?"),
					resource.TestCheckResourceAttr(resourceName, "csv_row_delimiter", "\\r\\n"),
					resource.TestCheckResourceAttr(resourceName, "data_format", "parquet"),
					resource.TestCheckResourceAttr(resourceName, "data_page_size", "1100000"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_delimiter", "UNDERSCORE"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_enabled", "true"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_sequence", "yyyymmddhh"),
					resource.TestCheckResourceAttr(resourceName, "date_partition_timezone", "America/Eastern"),
					resource.TestCheckResourceAttr(resourceName, "dict_page_size_limit", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "enable_statistics", "false"),
					resource.TestCheckResourceAttr(resourceName, "encoding_type", "plain"),
					resource.TestCheckResourceAttr(resourceName, "encryption_mode", "SSE_S3"),
					resource.TestCheckResourceAttr(resourceName, "external_table_definition", "etd"),
					resource.TestCheckResourceAttr(resourceName, "ignore_header_rows", "1"),
					resource.TestCheckResourceAttr(resourceName, "include_op_for_full_load", "true"),
					resource.TestCheckResourceAttr(resourceName, "max_file_size", "1000000"),
					resource.TestCheckResourceAttr(resourceName, "parquet_timestamp_in_millisecond", "true"),
					resource.TestCheckResourceAttr(resourceName, "parquet_version", "parquet-2-0"),
					resource.TestCheckResourceAttr(resourceName, "preserve_transactions", "false"),
					resource.TestCheckResourceAttr(resourceName, "rfc_4180", "false"),
					resource.TestCheckResourceAttr(resourceName, "row_group_length", "11000"),
					resource.TestCheckResourceAttrPair(resourceName, "service_access_role_arn", "aws_iam_role.iam_role", "arn"),
					resource.TestCheckResourceAttr(resourceName, "timestamp_column_name", "tx_commit_time"),
					resource.TestCheckResourceAttr(resourceName, "use_csv_no_sup_value", "true"),
					resource.TestCheckResourceAttr(resourceName, "use_task_start_time_for_full_load_timestamp", "true"),
				),
			},
		},
	})
}

func testAccS3EndpointConfig_basic(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_s3_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  add_column_name                             = true
  bucket_folder                               = "folder"
  bucket_name                                 = "bucket_name"
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = true
  cdc_max_batch_interval                      = 100
  cdc_min_file_size                           = 16
  cdc_path                                    = "cdc/path"
  compression_type                            = "GZIP"
  csv_delimiter                               = ";"
  csv_no_sup_value                            = "x"
  csv_null_value                              = "?"
  csv_row_delimiter                           = "\\r\\n"
  data_format                                 = "parquet"
  data_page_size                              = 1100000
  date_partition_delimiter                    = "UNDERSCORE"
  date_partition_enabled                      = true
  date_partition_sequence                     = "yyyymmddhh"
  date_partition_timezone                     = "Asia/Seoul"
  dict_page_size_limit                        = 1000000
  enable_statistics                           = false
  encoding_type                               = "plain"
  encryption_mode                             = "SSE_S3"
  external_table_definition                   = "etd"
  ignore_header_rows                          = 1
  include_op_for_full_load                    = true
  max_file_size                               = 1000000
  parquet_timestamp_in_millisecond            = true
  parquet_version                             = "parquet-2-0"
  preserve_transactions                       = false
  rfc_4180                                    = false
  row_group_length                            = 11000
  service_access_role_arn                     = aws_iam_role.iam_role.arn
  timestamp_column_name                       = "tx_commit_time"
  use_csv_no_sup_value                        = true
  use_task_start_time_for_full_load_timestamp = true

  depends_on = [aws_iam_role_policy.dms_s3_access]
}

resource "aws_iam_role" "iam_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = %[1]q
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, rName)
}

func testAccS3EndpointConfig_udpate(rName string) string {
	return fmt.Sprintf(`
data "aws_partition" "current" {}

resource "aws_dms_s3_endpoint" "test" {
  endpoint_id   = %[1]q
  endpoint_type = "target"
  ssl_mode      = "none"

  tags = {
    Name   = %[1]q
    Update = "to-update"
    Remove = "to-remove"
  }

  add_column_name                             = true
  bucket_folder                               = "folder"
  bucket_name                                 = "updated_name"
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = true
  cdc_max_batch_interval                      = 100
  cdc_min_file_size                           = 16
  cdc_path                                    = "cdc/path"
  compression_type                            = "GZIP"
  csv_delimiter                               = ";"
  csv_no_sup_value                            = "x"
  csv_null_value                              = "?"
  csv_row_delimiter                           = "\\r\\n"
  data_format                                 = "parquet"
  data_page_size                              = 1100000
  date_partition_delimiter                    = "SLASH"
  date_partition_enabled                      = true
  date_partition_sequence                     = "yyyymmddhh"
  date_partition_timezone                     = "America/Eastern"
  dict_page_size_limit                        = 1000000
  enable_statistics                           = false
  encoding_type                               = "plain"
  encryption_mode                             = "SSE_S3"
  external_table_definition                   = "etd"
  ignore_header_rows                          = 1
  include_op_for_full_load                    = true
  max_file_size                               = 1000000
  parquet_timestamp_in_millisecond            = true
  parquet_version                             = "parquet-2-0"
  preserve_transactions                       = false
  rfc_4180                                    = true
  row_group_length                            = 11000
  service_access_role_arn                     = aws_iam_role.iam_role.arn
  timestamp_column_name                       = "tx_commit_time"
  use_csv_no_sup_value                        = true
  use_task_start_time_for_full_load_timestamp = true

  depends_on = [aws_iam_role_policy.dms_s3_access]
}

resource "aws_iam_role" "iam_role" {
  name = %[1]q

  assume_role_policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Action": "sts:AssumeRole",
			"Principal": {
				"Service": "dms.${data.aws_partition.current.dns_suffix}"
			},
			"Effect": "Allow"
		}
	]
}
EOF
}

resource "aws_iam_role_policy" "dms_s3_access" {
  name = %[1]q
  role = aws_iam_role.iam_role.name

  policy = <<EOF
{
	"Version": "2012-10-17",
	"Statement": [
		{
			"Effect": "Allow",
			"Action": [
				"s3:CreateBucket",
				"s3:ListBucket",
				"s3:DeleteBucket",
				"s3:GetBucketLocation",
				"s3:GetObject",
				"s3:PutObject",
				"s3:DeleteObject",
				"s3:GetObjectVersion",
				"s3:GetBucketPolicy",
				"s3:PutBucketPolicy",
				"s3:DeleteBucketPolicy"
			],
			"Resource": "*"
		}
	]
}
EOF
}
`, rName)
}
