# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_mwaaserverless_workflow" "test" {
  name     = var.rName
  role_arn = aws_iam_role.test.arn

  definition_s3_location {
    bucket     = aws_s3_bucket.test.id
    object_key = aws_s3_object.test.key
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }
}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  bucket = aws_s3_bucket.test.id
  key    = "workflow.yaml"

  content = <<-YAML
    name: ${var.rName}
    schedule: "@daily"
    tasks:
      - id: start
        operator: airflow.operators.empty.EmptyOperator
  YAML
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [{
      Action = "sts:AssumeRole"
      Effect = "Allow"
      Principal = {
        Service = "airflow-serverless.amazonaws.com"
      }
    }]
  })
}

resource "null_resource" "test" {}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "unknownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagKey" {
  type     = string
  nullable = false
}

variable "knownTagValue" {
  type     = string
  nullable = false
}
