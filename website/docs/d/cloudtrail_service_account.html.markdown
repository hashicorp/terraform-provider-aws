---
subcategory: "CloudTrail"
layout: "aws"
page_title: "AWS: aws_cloudtrail_service_account"
description: |-
  Get AWS CloudTrail Service Account ID for storing trail data in S3.
---

# Data Source: aws_cloudtrail_service_account

Use this data source to get the Account ID of the [AWS CloudTrail Service Account](http://docs.aws.amazon.com/awscloudtrail/latest/userguide/cloudtrail-supported-regions.html)
in a given region for the purpose of allowing CloudTrail to store trail data in S3.

~> **Note:** AWS documentation [states that](https://docs.aws.amazon.com/awscloudtrail/latest/userguide/create-s3-bucket-policy-for-cloudtrail.html#troubleshooting-s3-bucket-policy) a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.

## Example Usage

```terraform
data "aws_cloudtrail_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-cloudtrail-logging-test-bucket"
  force_destroy = true
}

data "aws_iam_policy_document" "allow_cloudtrail_logging" {
  statement {
    sid    = "Put bucket policy needed for trails"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [data.aws_cloudtrail_service_account.main.arn]
    }

    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.bucket.arn}/*"]
  }

  statement {
    sid    = "Get bucket policy needed for trails"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [data.aws_cloudtrail_service_account.main.arn]
    }

    actions   = ["s3:GetBucketAcl"]
    resources = [aws_s3_bucket.bucket.arn]
  }
}

resource "aws_s3_bucket_policy" "allow_cloudtrail_logging" {
  bucket = aws_s3_bucket.bucket.id
  policy = data.aws_iam_policy_document.allow_cloudtrail_logging.json
}
```

## Argument Reference

* `region` - (Optional) Name of the region whose AWS CloudTrail account ID is desired.
Defaults to the region from the AWS provider configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS CloudTrail service account in the selected region.
* `arn` - ARN of the AWS CloudTrail service account in the selected region.
