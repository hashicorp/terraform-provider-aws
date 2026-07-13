# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_cloudwatch_log_delivery_source" "test" {
  name         = var.rName
  log_type     = "APPLICATION_LOGS"
  resource_arn = aws_bedrockagent_knowledge_base.test.arn
}

resource "aws_bedrockagent_knowledge_base" "test" {
  name     = var.rName
  role_arn = aws_iam_role.test.arn

  knowledge_base_configuration {
    type = "VECTOR"

    vector_knowledge_base_configuration {
      embedding_model_arn = "arn:${data.aws_partition.current.partition}:bedrock:${data.aws_region.current.region}::foundation-model/amazon.titan-embed-text-v2:0"
      embedding_model_configuration {
        bedrock_embedding_model_configuration {
          dimensions          = 256
          embedding_data_type = "FLOAT32"
        }
      }
    }
  }

  storage_configuration {
    type = "S3_VECTORS"

    s3_vectors_configuration {
      index_arn = aws_s3vectors_index.test.index_arn
    }
  }

  depends_on = [aws_iam_role_policy.test]
}

data "aws_region" "current" {
}
data "aws_partition" "current" {}

data "aws_iam_policy_document" "assume_role_bedrock" {
  statement {
    effect = "Allow"
    principals {
      type        = "Service"
      identifiers = ["bedrock.amazonaws.com"]
    }
    actions = ["sts:AssumeRole"]
  }
}

data "aws_iam_policy_document" "bedrock" {
  statement {
    effect    = "Allow"
    actions   = ["bedrock:InvokeModel"]
    resources = ["*"]
  }
  statement {
    effect    = "Allow"
    actions   = ["s3:ListBucket", "s3:GetObject"]
    resources = ["*"]
  }
  statement {
    effect = "Allow"
    actions = [
      "s3vectors:GetIndex",
      "s3vectors:QueryVectors",
      "s3vectors:PutVectors",
      "s3vectors:GetVectors",
      "s3vectors:DeleteVectors"
    ]
    resources = ["*"]
  }
}

resource "aws_iam_role" "test" {
  assume_role_policy = data.aws_iam_policy_document.assume_role_bedrock.json
  name               = var.rName
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.name
  policy = data.aws_iam_policy_document.bedrock.json
}

resource "aws_s3vectors_vector_bucket" "test" {
  vector_bucket_name = var.rName
  force_destroy      = true
}

resource "aws_s3vectors_index" "test" {
  index_name         = var.rName
  vector_bucket_name = aws_s3vectors_vector_bucket.test.vector_bucket_name

  data_type       = "float32"
  dimension       = 256
  distance_metric = "euclidean"
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "6.51.0"
    }
  }
}

provider "aws" {}
