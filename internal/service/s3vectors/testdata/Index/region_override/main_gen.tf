# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3vectors_vector_bucket" "test" {
  region = var.region

  vector_bucket_name = var.rName
}

resource "aws_s3vectors_index" "test" {
  region = var.region

  index_name         = var.rName
  vector_bucket_name = aws_s3vectors_vector_bucket.test.vector_bucket_name

  data_type       = "float32"
  dimension       = 2
  distance_metric = "euclidean"
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
