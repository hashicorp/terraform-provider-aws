# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3vectors_vector_bucket" "test" {
  region = var.region

  vector_bucket_name = var.rName
}

resource "aws_s3vectors_vector_bucket_policy" "test" {
  region = var.region

  vector_bucket_arn = aws_s3vectors_vector_bucket.test.vector_bucket_arn

  policy = <<EOF
{
  "Version": "2012-10-17",
  "Id": "writePolicy",
  "Statement": [{
    "Sid": "writeStatement",
    "Effect": "Allow",
    "Principal": {
      "AWS": "${data.aws_caller_identity.current.account_id}"
    },
    "Action": [
      "s3vectors:PutVectors"
    ],
    "Resource": "*"
  }]
}
EOF
}

data "aws_caller_identity" "current" {}
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
