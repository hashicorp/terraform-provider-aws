# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

resource "aws_macie2_classification_export_configuration" "test" {
  depends_on = [
    aws_macie2_account.test,
    aws_s3_bucket_policy.test,
  ]
  s3_destination {
    bucket_name = aws_s3_bucket.test.bucket
    kms_key_arn = aws_kms_key.test.arn
  }
}

resource "aws_macie2_account" "test" {}

resource "aws_s3_bucket" "test" {
  force_destroy = true
}

resource "aws_s3_bucket_policy" "test" {
  bucket = aws_s3_bucket.test.bucket
  policy = jsonencode(
    {
      "Version" : "2012-10-17",
      "Statement" : [
        {
          "Sid" : "Deny non-HTTPS access",
          "Effect" : "Deny",
          "Principal" : "*",
          "Action" : "s3:*",
          "Resource" : "${aws_s3_bucket.test.arn}/*",
          "Condition" : {
            "Bool" : {
              "aws:SecureTransport" : "false"
            }
          }
        },
        {
          "Sid" : "Allow Macie to upload objects to the bucket",
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "macie.${data.aws_partition.current.dns_suffix}"
          },
          "Action" : "s3:PutObject",
          "Resource" : "${aws_s3_bucket.test.arn}/*"
        },
        {
          "Sid" : "Allow Macie to use the getBucketLocation operation",
          "Effect" : "Allow",
          "Principal" : {
            "Service" : "macie.${data.aws_partition.current.dns_suffix}"
          },
          "Action" : "s3:GetBucketLocation",
          "Resource" : aws_s3_bucket.test.arn,
          "Condition" : {
            "StringEquals" : {
              "aws:SourceAccount" : data.aws_caller_identity.current.account_id
            },
            "ArnLike" : {
              "aws:SourceArn" : [
                "arn:${data.aws_partition.current.partition}:macie2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:export-configuration:*",
                "arn:${data.aws_partition.current.partition}:macie2:${data.aws_region.current.name}:${data.aws_caller_identity.current.account_id}:classification-job/*"
              ]
            }
          }
        }
      ]
    }
  )
}

resource "aws_kms_key" "test" {
  deletion_window_in_days = 7
  enable_key_rotation     = true

  policy = jsonencode({
    "Version" : "2012-10-17",
    "Id" : "allow_macie",
    "Statement" : [
      {
        "Sid" : "Allow Macie to use the key",
        "Effect" : "Allow",
        "Principal" : {
          "Service" : "macie.${data.aws_partition.current.dns_suffix}"
        },
        "Action" : [
          "kms:GenerateDataKey",
          "kms:Encrypt"
        ],
        "Resource" : "*"
      },
      {
        "Sid" : "Enable IAM User Permissions",
        "Effect" : "Allow",
        "Principal" : {
          "AWS" : "arn:${data.aws_partition.current.partition}:iam::${data.aws_caller_identity.current.account_id}:root"
        },
        "Action" : "kms:*",
        "Resource" : "*"
      }
    ]
  })
}

data "aws_caller_identity" "current" {}

data "aws_partition" "current" {}

data "aws_region" "current" {}

