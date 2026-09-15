# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3tables_table_bucket_replication" "test" {
  table_bucket_arn = aws_s3tables_table_bucket.source.arn
  role             = aws_iam_role.test.arn

  rule {
    destination {
      destination_table_bucket_arn = aws_s3tables_table_bucket.target.arn
    }
  }
}

data "aws_service_principal" "current" {
  service_name = "s3"
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Action": "sts:AssumeRole",
      "Principal": {
        "Service": "${data.aws_service_principal.current.name}"
      },
      "Effect": "Allow",
      "Sid": ""
    }
  ]
}
POLICY
}

resource "aws_s3tables_table_bucket" "source" {
  name = format("%[1]s-source", var.rName)
}

resource "aws_s3tables_table_bucket" "target" {
  name = format("%[1]s-target", var.rName)
}
variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
