# Copyright IBM Corp. 2014, 2025
# SPDX-License-Identifier: MPL-2.0

provider "null" {}

resource "aws_workspacesweb_session_logger" "test" {
  display_name = var.rName

  event_filter {
    all {}
  }

  log_configuration {
    s3 {
      bucket           = aws_s3_bucket.test.bucket
      folder_structure = "Flat"
      log_file_format  = "Json"
    }
  }

  tags = {
    (var.unknownTagKey) = null_resource.test.id
    (var.knownTagKey)   = var.knownTagValue
  }

  depends_on = [aws_s3_bucket_policy.allow_write_access]

}

resource "aws_s3_bucket" "test" {
  bucket        = var.rName
  force_destroy = true
}

data "aws_iam_policy_document" "allow_write_access" {
  statement {
    principals {
      type        = "Service"
      identifiers = ["workspaces-web.amazonaws.com"]
    }

    actions = [
      "s3:PutObject"
    ]

    resources = [
      aws_s3_bucket.test.arn,
      "${aws_s3_bucket.test.arn}/*"
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_write_access" {
  bucket = aws_s3_bucket.test.id
  policy = data.aws_iam_policy_document.allow_write_access.json
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
