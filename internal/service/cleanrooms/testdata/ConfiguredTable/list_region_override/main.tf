# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket" "test" {
  region = var.region

  count  = var.resource_count
  bucket = "${var.rName}-${count.index}"
}

resource "aws_glue_catalog_database" "test" {
  region = var.region

  count = var.resource_count
  name  = "${var.rName}-${count.index}"
}

resource "aws_glue_catalog_table" "test" {
  region = var.region

  count         = var.resource_count
  name          = "${var.rName}-${count.index}"
  database_name = aws_glue_catalog_database.test[count.index].name

  storage_descriptor {
    location = "s3://${aws_s3_bucket.test[count.index].bucket}"

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
  region = var.region

  count = var.resource_count

  name            = "${var.rName}-${count.index}"
  description     = "test description"
  analysis_method = "DIRECT_QUERY"
  allowed_columns = ["my_column_1", "my_column_2"]

  table_reference {
    database_name = aws_glue_catalog_database.test[count.index].name
    table_name    = aws_glue_catalog_table.test[count.index].name
  }

  depends_on = [aws_glue_catalog_table.test]
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}

variable "resource_count" {
  description = "Number of resources to create"
  type        = number
  nullable    = false
}

variable "region" {
  description = "Region to deploy resource in"
  type        = string
  nullable    = false
}
