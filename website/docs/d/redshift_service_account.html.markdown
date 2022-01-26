---
subcategory: "Redshift"
layout: "aws"
page_title: "AWS: aws_redshift_service_account"
description: |-
  Get AWS Redshift Service Account for storing audit data in S3.
---

# Data Source: aws_redshift_service_account

Use this data source to get the Account ID of the [AWS Redshift Service Account](http://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-enable-logging)
in a given region for the purpose of allowing Redshift to store audit data in S3.

## Example Usage

```terraform
data "aws_redshift_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-redshift-logging-test-bucket"
  force_destroy = true
}

data "aws_iam_policy_document" "audit_logging" {
  statement {
    sid = "Put bucket policy needed for audit logging"

    principals {
      type        = "AWS"
      identifiers = ["${data.aws_redshift_service_account.main.arn}"]
    }

    actions = [
      "s3:PutObject",
    ]

    effect = "Allow"

    resources = [
      "arn:aws:s3:::tf-redshift-logging-test-bucket/*",
    ]
  }

  statement {
    sid = "Get bucket policy needed for audit logging"

    principals {
      type        = "AWS"
      identifiers = ["${data.aws_redshift_service_account.main.arn}"]
    }

    actions = [
      "s3:GetBucketAcl",
    ]

    effect = "Allow"

    resources = [
      "arn:aws:s3:::tf-redshift-logging-test-bucket",
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_audit_logging" {
  bucket = aws_s3_bucket.bucket.id
  policy = data.aws_iam_policy_document.audit_logging.json
}
```

## Argument Reference

* `region` - (Optional) Name of the region whose AWS Redshift account ID is desired.
Defaults to the region from the AWS provider configuration.

## Attributes Reference

* `id` - The ID of the AWS Redshift service account in the selected region.
* `arn` - The ARN of the AWS Redshift service account in the selected region.
