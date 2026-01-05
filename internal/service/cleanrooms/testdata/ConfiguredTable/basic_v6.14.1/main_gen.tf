# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket" "test" {
  bucket = var.rName
}

resource "aws_glue_catalog_database" "test" {
  name = var.rName
}

resource "aws_glue_catalog_table" "test" {
  name          = var.rName
  database_name = var.rName

  storage_descriptor {
    location = "s3://${aws_s3_bucket.test.bucket}"

    columns {
      name = "my_column_1"
      type = "string"
    }

    columns {
      name = "my_column_2"
      type = "string"
    }
  }
}

resource "aws_cleanrooms_configured_table" "test" {
  name            = "test-name"
  description     = "test description"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = ["my_column_1", "my_column_2"]

  table_reference {
    database_name = var.rName
    table_name    = var.rName
  }

  depends_on = [aws_glue_catalog_table.test]
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
      version = "6.14.1"
    }
  }
}

provider "aws" {}
