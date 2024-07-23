// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package connect_test

import (
	"fmt"
	"testing"

	sdkacctest "github.com/hashicorp/terraform-plugin-testing/helper/acctest"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func testAccInstanceStorageConfigDataSource_KinesisFirehoseConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_instance_storage_config.test"
	datasourceName := "data.aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigDataSourceConfig_kinesisFirehoseConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrResourceType, resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.#", resourceName, "storage_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_firehose_config.#", resourceName, "storage_config.0.kinesis_firehose_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_firehose_config.0.firehose_arn", resourceName, "storage_config.0.kinesis_firehose_config.0.firehose_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.storage_type", resourceName, "storage_config.0.storage_type"),
				),
			},
		},
	})
}

func testAccInstanceStorageConfigDataSource_KinesisStreamConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_instance_storage_config.test"
	datasourceName := "data.aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigDataSourceConfig_kinesisStreamConfig(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrResourceType, resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.#", resourceName, "storage_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_stream_config.#", resourceName, "storage_config.0.kinesis_stream_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_stream_config.0.stream_arn", resourceName, "storage_config.0.kinesis_stream_config.0.stream_arn"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.storage_type", resourceName, "storage_config.0.storage_type"),
				),
			},
		},
	})
}

func testAccInstanceStorageConfigDataSource_KinesisVideoStreamConfig(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_instance_storage_config.test"
	datasourceName := "data.aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccInstanceStorageConfigDataSourceConfig_kinesisVideoStreamConfig(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrResourceType, resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.#", resourceName, "storage_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.#", resourceName, "storage_config.0.kinesis_video_stream_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.0.prefix", resourceName, "storage_config.0.kinesis_video_stream_config.0.prefix"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.0.retention_period_hours", resourceName, "storage_config.0.kinesis_video_stream_config.0.retention_period_hours"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.#", resourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.0.encryption_type", resourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.0.encryption_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.0.key_id", resourceName, "storage_config.0.kinesis_video_stream_config.0.encryption_config.0.key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.storage_type", resourceName, "storage_config.0.storage_type"),
				),
			},
		},
	})
}

func testAccInstanceStorageConfigDataSource_S3Config(t *testing.T) {
	ctx := acctest.Context(t)
	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	rName2 := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	resourceName := "aws_connect_instance_storage_config.test"
	datasourceName := "data.aws_connect_instance_storage_config.test"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ConnectServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{ // nosemgrep:ci.test-config-funcs-correct-form
				Config: testAccInstanceStorageConfigDataSourceConfig_S3Config(rName, rName2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrAssociationID, resourceName, names.AttrAssociationID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrInstanceID, resourceName, names.AttrInstanceID),
					resource.TestCheckResourceAttrPair(datasourceName, names.AttrResourceType, resourceName, names.AttrResourceType),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.#", resourceName, "storage_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.#", resourceName, "storage_config.0.s3_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.bucket_name", resourceName, "storage_config.0.s3_config.0.bucket_name"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.bucket_prefix", resourceName, "storage_config.0.s3_config.0.bucket_prefix"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.#", resourceName, "storage_config.0.s3_config.0.encryption_config.#"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type", resourceName, "storage_config.0.s3_config.0.encryption_config.0.encryption_type"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id", resourceName, "storage_config.0.s3_config.0.encryption_config.0.key_id"),
					resource.TestCheckResourceAttrPair(datasourceName, "storage_config.0.storage_type", resourceName, "storage_config.0.storage_type"),
				),
			},
		},
	})
}

func testAccInstanceStorageConfigDataSourceConfig_base(rName string) string {
	return fmt.Sprintf(`
resource "aws_connect_instance" "test" {
  identity_management_type = "CONNECT_MANAGED"
  inbound_calls_enabled    = true
  instance_alias           = %[1]q
  outbound_calls_enabled   = true
}
`, rName)
}

func testAccInstanceStorageConfigDataSourceConfig_kinesisFirehoseConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigDataSourceConfig_base(rName),
		fmt.Sprintf(`
data "aws_caller_identity" "current" {}
data "aws_partition" "current" {}

resource "aws_iam_role" "firehose" {
  name = %[1]q

  assume_role_policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Principal": {
        "Service": "firehose.amazonaws.com"
      },
      "Action": "sts:AssumeRole",
      "Condition": {
        "StringEquals": {
          "sts:ExternalId": "${data.aws_caller_identity.current.account_id}"
        }
      }
    }
  ]
}
EOF
}

resource "aws_s3_bucket" "bucket" {
  bucket = %[1]q
}

resource "aws_iam_role_policy" "firehose" {
  name = %[1]q
  role = aws_iam_role.firehose.id

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Sid": "",
      "Effect": "Allow",
      "Action": [
        "s3:AbortMultipartUpload",
        "s3:GetBucketLocation",
        "s3:GetObject",
        "s3:ListBucket",
        "s3:ListBucketMultipartUploads",
        "s3:PutObject"
      ],
      "Resource": [
        "${aws_s3_bucket.bucket.arn}",
        "${aws_s3_bucket.bucket.arn}/*"
      ]
    },
    {
      "Sid": "GlueAccess",
      "Effect": "Allow",
      "Action": [
        "glue:GetTable",
        "glue:GetTableVersion",
        "glue:GetTableVersions"
      ],
      "Resource": [
        "*"
      ]
    },
    {
      "Sid": "LakeFormationDataAccess",
      "Effect": "Allow",
      "Action": [
        "lakeformation:GetDataAccess"
      ],
      "Resource": "*"
    }
  ]
}
EOF
}

resource "aws_kinesis_firehose_delivery_stream" "test" {
  depends_on  = [aws_iam_role_policy.firehose]
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    role_arn   = aws_iam_role.firehose.arn
    bucket_arn = aws_s3_bucket.bucket.arn
  }
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "CONTACT_TRACE_RECORDS"

  storage_config {
    kinesis_firehose_config {
      firehose_arn = aws_kinesis_firehose_delivery_stream.test.arn
    }
    storage_type = "KINESIS_FIREHOSE"
  }
}

data "aws_connect_instance_storage_config" "test" {
  association_id = aws_connect_instance_storage_config.test.association_id
  instance_id    = aws_connect_instance.test.id
  resource_type  = aws_connect_instance_storage_config.test.resource_type
}
`, rName))
}

func testAccInstanceStorageConfigDataSourceConfig_kinesisStreamConfig(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kinesis_stream" "test" {
  name        = %[1]q
  shard_count = 2
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "CONTACT_TRACE_RECORDS"

  storage_config {
    kinesis_stream_config {
      stream_arn = aws_kinesis_stream.test.arn
    }
    storage_type = "KINESIS_STREAM"
  }
}

data "aws_connect_instance_storage_config" "test" {
  association_id = aws_connect_instance_storage_config.test.association_id
  instance_id    = aws_connect_instance.test.id
  resource_type  = aws_connect_instance_storage_config.test.resource_type
}
`, rName2))
}

func testAccInstanceStorageConfigDataSourceConfig_kinesisVideoStreamConfig(rName string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigDataSourceConfig_base(rName),
		`
resource "aws_kms_key" "test" {
  description             = "KMS Key"
  deletion_window_in_days = 10
}

resource "aws_connect_instance_storage_config" "test" {
  instance_id   = aws_connect_instance.test.id
  resource_type = "MEDIA_STREAMS"

  storage_config {
    kinesis_video_stream_config {
      prefix                 = "tf-test-prefix"
      retention_period_hours = 1

      encryption_config {
        encryption_type = "KMS"
        key_id          = aws_kms_key.test.arn
      }
    }
    storage_type = "KINESIS_VIDEO_STREAM"
  }
}

data "aws_connect_instance_storage_config" "test" {
  association_id = aws_connect_instance_storage_config.test.association_id
  instance_id    = aws_connect_instance.test.id
  resource_type  = aws_connect_instance_storage_config.test.resource_type
}
`)
}

func testAccInstanceStorageConfigDataSourceConfig_S3Config(rName, rName2 string) string {
	return acctest.ConfigCompose(
		testAccInstanceStorageConfigDataSourceConfig_base(rName),
		fmt.Sprintf(`
resource "aws_kms_key" "test" {
  description             = "KMS Key for Bucket"
  deletion_window_in_days = 10
}

resource "aws_s3_bucket" "test" {
  bucket = %[1]q
}

resource "aws_s3_bucket_server_side_encryption_configuration" "test" {
  bucket = aws_s3_bucket.test.bucket

  rule {
    apply_server_side_encryption_by_default {
      kms_master_key_id = aws_kms_key.test.arn
      sse_algorithm     = "aws:kms"
    }
  }
}

resource "aws_connect_instance_storage_config" "test" {
  depends_on = [aws_s3_bucket_server_side_encryption_configuration.test]

  instance_id   = aws_connect_instance.test.id
  resource_type = "CHAT_TRANSCRIPTS"

  storage_config {
    s3_config {
      bucket_name   = aws_s3_bucket.test.id
      bucket_prefix = "tf-test-Chat-Transcripts"

      encryption_config {
        encryption_type = "KMS"
        key_id          = aws_kms_key.test.arn
      }
    }
    storage_type = "S3"
  }
}

data "aws_connect_instance_storage_config" "test" {
  association_id = aws_connect_instance_storage_config.test.association_id
  instance_id    = aws_connect_instance.test.id
  resource_type  = aws_connect_instance_storage_config.test.resource_type
}
`, rName2))
}
