# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_glue_catalog_table" "test" {
  region = var.region

  name          = var.rName
  database_name = aws_glue_catalog_database.test.name
}

resource "aws_glue_catalog_database" "test" {
  region = var.region

  name = var.rName
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
