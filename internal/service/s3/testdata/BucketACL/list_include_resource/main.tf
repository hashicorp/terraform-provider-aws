# Copyright IBM Corp. 2014, 2026
# SPDX-License-Identifier: MPL-2.0

resource "aws_s3_bucket_acl" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].bucket

  access_control_policy {
    grant {
      grantee {
        id   = data.aws_canonical_user_id.current.id
        type = "CanonicalUser"
      }
      permission = "FULL_CONTROL"
    }

    owner {
      id = data.aws_canonical_user_id.current.id
    }
  }

  depends_on = [aws_s3_bucket_ownership_controls.test]
}

resource "aws_s3_bucket" "test" {
  count = var.resource_count

  bucket = "${var.rName}-${count.index}"
}

resource "aws_s3_bucket_ownership_controls" "test" {
  count = var.resource_count

  bucket = aws_s3_bucket.test[count.index].bucket

  rule {
    object_ownership = "BucketOwnerPreferred"
  }
}

data "aws_canonical_user_id" "current" {}

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
