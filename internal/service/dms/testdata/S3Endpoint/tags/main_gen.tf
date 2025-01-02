# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_dms_s3_endpoint" "test" {
  endpoint_id   = var.rName
  endpoint_type = "target"
  ssl_mode      = "none"

  add_column_name                             = true
  add_trailing_padding_character              = false
  bucket_folder                               = "folder"
  bucket_name                                 = "bucket_name"
  canned_acl_for_objects                      = "private"
  cdc_inserts_and_updates                     = true
  cdc_inserts_only                            = false
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
  expected_bucket_owner                       = data.aws_caller_identity.current.account_id
  ignore_header_rows                          = 1
  include_op_for_full_load                    = true
  max_file_size                               = 1000000
  parquet_timestamp_in_millisecond            = true
  parquet_version                             = "parquet-2-0"
  preserve_transactions                       = false
  rfc_4180                                    = false
  row_group_length                            = 11000
  server_side_encryption_kms_key_id           = aws_kms_key.test.arn
  service_access_role_arn                     = aws_iam_role.test.arn
  timestamp_column_name                       = "tx_commit_time"
  use_csv_no_sup_value                        = false
  use_task_start_time_for_full_load_timestamp = true
  glue_catalog_generation                     = true

  depends_on = [aws_iam_role_policy.test]

  tags = var.resource_tags
}

resource "aws_kms_key" "test" {
  description = var.rName

  policy = jsonencode({
    Version = "2012-10-17"
    Id      = var.rName
    Statement = [{
      Sid    = var.rName
      Effect = "Allow"
      Principal = {
        AWS = "*"
      }
      Action   = "kms:*"
      Resource = "*"
    }]
  })
}

# testAccS3EndpointConfig_base

data "aws_partition" "current" {}
data "aws_caller_identity" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "dms.${data.aws_partition.current.dns_suffix}"
      }
    }]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.name

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Effect = "Allow"
      Action = [
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
      ]
      Resource = "*"
    }]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_tags" {
  description = "Tags to set on resource. To specify no tags, set to `null`"
  # Not setting a default, so that this must explicitly be set to `null` to specify no tags
  type     = map(string)
  nullable = true
}
