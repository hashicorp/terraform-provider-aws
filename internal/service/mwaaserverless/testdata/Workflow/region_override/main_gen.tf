# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_mwaaserverless_workflow" "test" {
  region = var.region

  name     = var.rName
  role_arn = aws_iam_role.test.arn

  definition_s3_location {
    bucket     = aws_s3_bucket.test.id
    object_key = aws_s3_object.test.key
  }
}

resource "aws_s3_bucket" "test" {
  region = var.region

  bucket        = var.rName
  force_destroy = true
}

resource "aws_s3_object" "test" {
  region = var.region

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

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
