# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

terraform {
  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "0.0.0-dev"
    }
  }
}

provider "aws" {
  region = var.region
}

variable "region" {
  type = string
}

variable "target_arn" {
  description = "SageMaker Lakehouse target ARN"
  type        = string
}

resource "aws_dynamodb_table" "this" {
  name           = "example"
  hash_key       = "pk"
  read_capacity  = 1
  write_capacity = 1

  attribute {
    name = "pk"
    type = "S"
  }

  point_in_time_recovery { enabled = true }
}

resource "aws_glue_inbound_integration" "this" {
  integration_name = "example"
  source_arn       = aws_dynamodb_table.this.arn
  target_arn       = var.target_arn
}


