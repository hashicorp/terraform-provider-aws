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

~> **Note:** AWS documentation [states that](https://docs.aws.amazon.com/redshift/latest/mgmt/db-auditing.html#db-auditing-bucket-permissions) a [service principal name](https://docs.aws.amazon.com/IAM/latest/UserGuide/reference_policies_elements_principal.html#principal-services) should be used instead of an AWS account ID in any relevant IAM policy.
The `aws_redshift_service_account` data source has been deprecated and will be removed in a future version.

## Example Usage

```terraform
data "aws_redshift_service_account" "main" {}

resource "aws_s3_bucket" "bucket" {
  bucket        = "tf-redshift-logging-test-bucket"
  force_destroy = true
}

data "aws_iam_policy_document" "allow_audit_logging" {
  statement {
    sid    = "Put bucket policy needed for audit logging"
    effect = "Allow"

    principals {
      type        = "AWS"
      identifiers = [data.aws_redshift_service_account.main.arn]
    }

    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.bucket.arn}/*"]
  }

  statement {
    sid    = "Get bucket policy needed for audit logging"
    effect = "Allow"

    principals {
      type = "AWS"
      identifiers = [
        data.aws_redshift_service_account.main.arn,
      ]
    }

    actions   = ["s3:GetBucketAcl"]
    resources = data.aws_s3_bucket.bucket.arn
  }
}

resource "aws_s3_bucket_policy" "allow_audit_logging" {
  bucket = aws_s3_bucket.bucket.id
  policy = data.aws_iam_policy_document.allow_audit_logging.json
}
```

## Argument Reference

* `region` - (Optional) Name of the region whose AWS Redshift account ID is desired.
Defaults to the region from the AWS provider configuration.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - ID of the AWS Redshift service account in the selected region.
* `arn` - ARN of the AWS Redshift service account in the selected region.
