resource "aws_msk_channel" "test" {
{{- template "region" }}
  channel_name = var.rName
  cluster_arn  = aws_msk_cluster.test.arn

  topic_configuration {
    topic_arn = aws_msk_topic.test.arn

    record_converter {
      value_converter = "BYTE_ARRAY"
    }
  }

  s3_destination {
    service_execution_role_arn = aws_iam_role.test.arn

    dead_letter_queue_s3 {
      bucket_arn = aws_s3_bucket.dlq.arn
    }

    storage {
      bucket_arn       = aws_s3_bucket.test.arn
      compression_type = "NONE"
      storage_class    = "STANDARD"
    }
  }

  depends_on = [aws_iam_role_policy.test]
{{- template "tags" . }}
}

resource "aws_msk_topic" "test" {
{{- template "region" }}
  name               = var.rName
  cluster_arn        = aws_msk_cluster.test.arn
  partition_count    = 2
  replication_factor = 3
}

resource "aws_msk_cluster" "test" {
{{- template "region" }}
  cluster_name           = var.rName
  kafka_version          = "3.8.x"
  number_of_broker_nodes = 3

  broker_node_group_info {
    client_subnets  = aws_subnet.test[*].id
    instance_type   = "express.m7g.large"
    security_groups = [aws_security_group.test.id]
  }
}

{{ template "acctest.ConfigVPCWithSubnets" 3 }}

resource "aws_security_group" "test" {
{{- template "region" }}
  vpc_id = aws_vpc.test.id
}

resource "aws_s3_bucket" "test" {
{{- template "region" }}
  bucket        = "${var.rName}-dest"
  force_destroy = true
}

resource "aws_s3_bucket" "dlq" {
{{- template "region" }}
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
        Sid    = "DeliveryBucketList"
        Effect = "Allow"
        Action = [
          "s3:ListBucket",
          "s3:ListBucketMultipartUploads",
          "s3:GetBucketLocation",
        ]
        Resource = [
          aws_s3_bucket.test.arn,
          "${aws_s3_bucket.test.arn}/*",
        ]
      },
      {
        Sid    = "DeliveryBucketWrite"
        Effect = "Allow"
        Action = [
          "s3:UploadPart",
          "s3:CompleteMultipartUpload",
          "s3:CreateMultipartUpload",
          "s3:PutObject",
          "s3:ListMultipartUploads",
          "s3:ListMultipartUploadParts",
        ]
        Resource = ["${aws_s3_bucket.test.arn}/*"]
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
    ]
  })
}
