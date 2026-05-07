// Copyright IBM Corp. 2014, 2026
// SPDX-License-Identifier: MPL-2.0

package elasticache_test

import (
	"fmt"
	"testing"

	awstypes "github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-provider-aws/internal/acctest"
	"github.com/hashicorp/terraform-provider-aws/names"
)

func TestAccElastiCacheClusterDataSource_basic(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	resourceName := "aws_elasticache_cluster.test"
	dataSourceName := "data.aws_elasticache_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterDataSourceConfig_basic(rName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrAvailabilityZone, resourceName, names.AttrAvailabilityZone),
					resource.TestCheckResourceAttrPair(dataSourceName, "cluster_address", resourceName, "cluster_address"),
					resource.TestCheckResourceAttrPair(dataSourceName, "configuration_endpoint", resourceName, "configuration_endpoint"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrEngine, resourceName, names.AttrEngine),
					resource.TestCheckResourceAttrPair(dataSourceName, "ip_discovery", resourceName, "ip_discovery"),
					resource.TestCheckResourceAttrPair(dataSourceName, "network_type", resourceName, "network_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "node_type", resourceName, "node_type"),
					resource.TestCheckResourceAttrPair(dataSourceName, "num_cache_nodes", resourceName, "num_cache_nodes"),
					resource.TestCheckResourceAttrPair(dataSourceName, names.AttrPort, resourceName, names.AttrPort),
					resource.TestCheckResourceAttrPair(dataSourceName, "preferred_outpost_arn", resourceName, "preferred_outpost_arn"),
				),
			},
		},
	})
}

func testAccClusterDataSourceConfig_basic(rName string) string {
	return fmt.Sprintf(`
resource "aws_elasticache_cluster" "test" {
  cluster_id      = %[1]q
  engine          = "memcached"
  node_type       = "cache.t3.small"
  num_cache_nodes = 1
  port            = 11211
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName)
}

func TestAccElastiCacheClusterDataSource_Engine_Redis_LogDeliveryConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := acctest.RandomWithPrefix(t, acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_cluster.test"

	acctest.ParallelTest(ctx, t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dataSourceEngineRedisLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "redis"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
		},
	})
}

func testAccClusterConfig_dataSourceEngineValkeyLogDeliveryConfigurations(rName string, slowLogDeliveryEnabled bool, slowDeliveryDestination awstypes.DestinationType, slowDeliveryFormat awstypes.LogFormat, engineLogDeliveryEnabled bool, engineDeliveryDestination awstypes.DestinationType, engineLogDeliveryFormat awstypes.LogFormat) string {
	return fmt.Sprintf(`
data "aws_iam_policy_document" "p" {
  statement {
    actions = [
      "logs:CreateLogStream",
      "logs:PutLogEvents"
    ]
    resources = ["${aws_cloudwatch_log_group.lg.arn}:log-stream:*"]
    principals {
      identifiers = ["delivery.logs.amazonaws.com"]
      type        = "Service"
    }
  }
}

resource "aws_cloudwatch_log_resource_policy" "rp" {
  policy_document = data.aws_iam_policy_document.p.json
  policy_name     = %[1]q
  depends_on = [
    aws_cloudwatch_log_group.lg
  ]
}

resource "aws_cloudwatch_log_group" "lg" {
  retention_in_days = 1
  name              = %[1]q
}

resource "aws_s3_bucket" "b" {
  force_destroy = true
}

resource "aws_iam_role" "r" {
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "firehose.amazonaws.com"
        }
      },
    ]
  })
  inline_policy {
    name = "my_inline_s3_policy"
    policy = jsonencode({
      Version = "2012-10-17"
      Statement = [
        {
          Action = [
            "s3:AbortMultipartUpload",
            "s3:GetBucketLocation",
            "s3:GetObject",
            "s3:ListBucket",
            "s3:ListBucketMultipartUploads",
            "s3:PutObject",
            "s3:PutObjectAcl",
          ]
          Effect   = "Allow"
          Resource = [aws_s3_bucket.b.arn, "${aws_s3_bucket.b.arn}/*"]
        },
      ]
    })
  }
}

resource "aws_kinesis_firehose_delivery_stream" "ds" {
  name        = %[1]q
  destination = "extended_s3"

  extended_s3_configuration {
    bucket_arn = aws_s3_bucket.b.arn
    role_arn   = aws_iam_role.r.arn
  }

  lifecycle {
    ignore_changes = [
      tags["LogDeliveryEnabled"],
    ]
  }
}

resource "aws_elasticache_cluster" "test" {
  cluster_id        = %[1]q
  engine            = "valkey"
  node_type         = "cache.t3.micro"
  num_cache_nodes   = 1
  port              = 6379
  apply_immediately = true
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[2]t") ? [""] : []
    content {
      destination      = (%[3]q == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : ((%[3]q == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = %[3]q
      log_format       = %[4]q
      log_type         = "slow-log"
    }
  }
  dynamic "log_delivery_configuration" {
    for_each = tobool("%[5]t") ? [""] : []
    content {
      destination      = (%[6]q == "cloudwatch-logs") ? aws_cloudwatch_log_group.lg.name : ((%[6]q == "kinesis-firehose") ? aws_kinesis_firehose_delivery_stream.ds.name : null)
      destination_type = %[6]q
      log_format       = %[7]q
      log_type         = "engine-log"
    }
  }
}

data "aws_elasticache_cluster" "test" {
  cluster_id = aws_elasticache_cluster.test.cluster_id
}
`, rName, slowLogDeliveryEnabled, slowDeliveryDestination, slowDeliveryFormat, engineLogDeliveryEnabled, engineDeliveryDestination, engineLogDeliveryFormat)
}

func TestAccElastiCacheClusterDataSource_Engine_Valkey_LogDeliveryConfigurations(t *testing.T) {
	ctx := acctest.Context(t)
	if testing.Short() {
		t.Skip("skipping long-running test in short mode")
	}

	rName := sdkacctest.RandomWithPrefix(acctest.ResourcePrefix)
	dataSourceName := "data.aws_elasticache_cluster.test"

	resource.ParallelTest(t, resource.TestCase{
		PreCheck:                 func() { acctest.PreCheck(ctx, t) },
		ErrorCheck:               acctest.ErrorCheck(t, names.ElastiCacheServiceID),
		ProtoV5ProviderFactories: acctest.ProtoV5ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccClusterConfig_dataSourceEngineValkeyLogDeliveryConfigurations(rName, true, awstypes.DestinationTypeKinesisFirehose, awstypes.LogFormatJson, true, awstypes.DestinationTypeCloudWatchLogs, awstypes.LogFormatText),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr(dataSourceName, names.AttrEngine, "valkey"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.#", "2"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.destination_type", "cloudwatch-logs"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_format", "text"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.0.log_type", "engine-log"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination", rName),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.destination_type", "kinesis-firehose"),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_format", names.AttrJSON),
					resource.TestCheckResourceAttr(dataSourceName, "log_delivery_configuration.1.log_type", "slow-log"),
				),
			},
		},
	})
}
