---
subcategory: ""
layout: "aws"
page_title: "AWS: aws_billing_service_account"
description: |-
  Get AWS Billing Service Account
---

# Data Source: aws_billing_service_account

Use this data source to get the Account ID of the [AWS Billing and Cost Management Service Account](http://docs.aws.amazon.com/awsaccountbilling/latest/aboutv2/billing-getting-started.html#step-2) for the purpose of permitting in S3 bucket policy.

## Example Usage

```terraform
data "aws_billing_service_account" "main" {}

resource "aws_s3_bucket" "billing_logs" {
  bucket = "my-billing-tf-test-bucket"
  acl    = "private"
}

data "aws_iam_policy_document" "billing_logging" {
  statement {
    principals {
      type        = "AWS"
      identifiers = ["${data.aws_billing_service_account.main.arn}"]
    }

    actions = [
      "s3:GetBucketAcl",
      "s3:GetBucketPolicy",
    ]

    effect = "Allow"

    resources = [
      "arn:aws:s3:::my-billing-tf-test-bucket",
    ]
  }

  statement {
    principals {
      type        = "AWS"
      identifiers = ["${data.aws_billing_service_account.main.arn}"]
    }

    actions = [
      "s3:PutObject",
    ]

    effect = "Allow"

    resources = [
      "arn:aws:s3:::my-billing-tf-test-bucket/*",
    ]
  }
}

resource "aws_s3_bucket_policy" "allow_billing_logging" {
  bucket = aws_s3_bucket.billing_logs.id
  policy = data.aws_iam_policy_document.billing_logging.json
}
```

## Attributes Reference

* `id` - The ID of the AWS billing service account.
* `arn` - The ARN of the AWS billing service account.
