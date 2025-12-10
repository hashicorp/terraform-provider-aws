# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_datasync_location_s3" "test" {
  s3_bucket_arn = aws_s3_bucket.test.arn
  subdirectory  = "/test"

  s3_config {
    bucket_access_role_arn = aws_iam_role.test.arn
  }
}

# testAccLocationS3Config_base

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

resource "aws_iam_role" "test" {
  name = var.rName

  assume_role_policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {
        "Service": "datasync.amazonaws.com"
      },
      "Action": "sts:AssumeRole"
    }
  ]
}
POLICY
}

resource "aws_iam_role_policy" "test" {
  role   = aws_iam_role.test.id
  policy = <<POLICY
{
  "Version": "2012-10-17",
  "Statement": [{
    "Action": [
      "s3:*"
    ],
    "Effect": "Allow",
    "Resource": [
      "${aws_s3_bucket.test.arn}",
      "${aws_s3_bucket.test.arn}/*"
    ]
  }]
}
POLICY
}

variable "rName" {
  description = "Name for resource"
  type        = string
  nullable    = false
}
