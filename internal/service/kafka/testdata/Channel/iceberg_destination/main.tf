# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_msk_channel" "test" {
  channel_name = var.rName
  cluster_arn  = aws_msk_cluster.test.arn

  topic_configuration {
    topic_arn = aws_msk_topic.test.arn

    record_converter {
      value_converter = "JSON"
    }

    record_schema {
      gsr_arn = aws_glue_schema.test.arn
    }
  }

  iceberg_destination {
    append_only                = true
    data_freshness_in_seconds  = 300
    service_execution_role_arn = aws_iam_role.test.arn

    catalog {
      warehouse_location = aws_s3tables_table_bucket.test.arn
    }

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    destination_table {
      destination_database_name = "example_namespace"
      destination_table_name    = "example_table"

      partition_spec {
        partition_strategy = "TIME_HOUR"

        source {
          source_name = "event_time"
        }
      }
    }

    schema_evolution {
      enable_schema_evolution = false
    }

    table_creation {
      enable_table_creation = true
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

resource "aws_msk_topic" "test" {
  name               = "${var.rName}-topic"
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 3
}

resource "aws_msk_cluster" "test" {
  cluster_name           = "${var.rName}-cluster"
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "express.m7g.large"
    security_groups = [aws_security_group.test.id]
  }
}

resource "aws_s3tables_table_bucket" "test" {
  name = var.rName
}

resource "aws_glue_registry" "test" {
  registry_name = var.rName
}

resource "aws_glue_schema" "test" {
  schema_name   = var.rName
  registry_arn  = aws_glue_registry.test.arn
  data_format   = "JSON"
  compatibility = "NONE"

  schema_definition = jsonencode({
    "$schema" = "http://json-schema.org/draft-07/schema#"
    type      = "object"
    properties = {
      id         = { type = "integer" }
      event_time = { type = "string", format = "date-time" }
    }
    required = ["id", "event_time"]
  })
}

# acctest.ConfigVPCWithSubnets(rName, 3)

resource "aws_vpc" "test" {
  cidr_block = "10.0.0.0/16"
}

# acctest.ConfigSubnets(rName, 3)

resource "aws_subnet" "test" {
  count = 3

  vpc_id            = aws_vpc.test.id
  availability_zone = data.aws_availability_zones.available.names[count.index]
  cidr_block        = cidrsubnet(aws_vpc.test.cidr_block, 8, count.index)
}

# acctest.ConfigAvailableAZsNoOptInDefaultExclude

data "aws_availability_zones" "available" {
  exclude_zone_ids = local.default_exclude_zone_ids
  state            = "available"

  filter {
    name   = "opt-in-status"
    values = ["opt-in-not-required"]
  }
}

locals {
  default_exclude_zone_ids = ["usw2-az4", "usgw1-az2"]
}

resource "aws_security_group" "test" {
  name   = var.rName
  vpc_id = aws_vpc.test.id
}

resource "aws_s3_bucket" "dlq" {
  bucket        = "${var.rName}-dlq"
  force_destroy = true
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Effect = "Allow"
        Principal = {
          Service = "kafka.amazonaws.com"
        }
        Action = "sts:AssumeRole"
        Condition = {
          StringEquals = {
            "aws:SourceAccount" = data.aws_caller_identity.current.account_id
          }
          ArnLike = {
            "aws:SourceArn" = "arn:${data.aws_partition.current.partition}:kafka:*:${data.aws_caller_identity.current.account_id}:channel/*"
          }
        }
      }
    ]
  })
}

resource "aws_iam_role_policy" "test" {
  name = var.rName
  role = aws_iam_role.test.id

  policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Sid    = "AllowS3TablesActions"
        Effect = "Allow"
        Action = [
          "s3tables:GetTable",
          "s3tables:GetTableMetadataLocation",
          "s3tables:UpdateTableMetadataLocation",
          "s3tables:CreateTable",
          "s3tables:PutTableData",
          "s3tables:CreateNamespace",
          "s3tables:GetTableData",
          "s3tables:GetTableBucket",
          "s3tables:TagResource",
          "s3tables:PutTableRecordExpirationConfiguration",
        ]
        Resource = [
          aws_s3tables_table_bucket.test.arn,
          "${aws_s3tables_table_bucket.test.arn}/table/*",
        ]
      },
      {
        Sid    = "DLQBucketAccess"
        Effect = "Allow"
        Action = [
          "s3:GetBucketLocation",
          "s3:PutObject",
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
        ]
        Resource = [
          aws_s3_bucket.dlq.arn,
          "${aws_s3_bucket.dlq.arn}/*",
        ]
      },
      {
        Sid    = "GlueSchemaRegistryAccess"
        Effect = "Allow"
        Action = ["glue:GetSchemaVersion"]
        Resource = [
          aws_glue_registry.test.arn,
          aws_glue_schema.test.arn,
        ]
      },
    ]
  })
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
