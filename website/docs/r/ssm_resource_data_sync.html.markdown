---
subcategory: "SSM (Systems Manager)"
layout: "aws"
page_title: "AWS: aws_ssm_resource_data_sync"
description: |-
  Provides a SSM resource data sync.
---

# Resource: aws_ssm_resource_data_sync

Provides a SSM resource data sync.

## Example Usage

```terraform
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test-bucket-1234"
}

data "aws_iam_policy_document" "hoge" {
  statement {
    sid    = "SSMBucketPermissionsCheck"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ssm.amazonaws.com"]
    }

    actions   = ["s3:GetBucketAcl"]
    resources = ["arn:aws:s3:::tf-test-bucket-1234"]
  }

  statement {
    sid    = "SSMBucketDelivery"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ssm.amazonaws.com"]
    }

    actions   = ["s3:PutObject"]
    resources = ["arn:aws:s3:::tf-test-bucket-1234/*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"
      values   = ["bucket-owner-full-control"]
    }
  }
}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = aws_s3_bucket.hoge.id
  policy = data.aws_iam_policy_document.hoge.json
}

resource "aws_ssm_resource_data_sync" "foo" {
  name = "foo"

  s3_destination {
    bucket_name = aws_s3_bucket.hoge.bucket
    region      = aws_s3_bucket.hoge.region
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required) Name for the configuration.
* `s3_destination` - (Required) Amazon S3 configuration details for the sync.

## s3_destination

`s3_destination` supports the following:

* `bucket_name` - (Required) Name of S3 bucket where the aggregated data is stored.
* `region` - (Required) Region with the bucket targeted by the Resource Data Sync.
* `kms_key_arn` - (Optional) ARN of an encryption key for a destination in Amazon S3.
* `prefix` - (Optional) Prefix for the bucket.
* `sync_format` - (Optional) A supported sync format. Only JsonSerDe is currently supported. Defaults to JsonSerDe.

## Attribute Reference

This resource exports no additional attributes.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import SSM resource data sync using the `name`. For example:

```terraform
import {
  to = aws_ssm_resource_data_sync.example
  id = "example-name"
}
```

Using `terraform import`, import SSM resource data sync using the `name`. For example:

```console
% terraform import aws_ssm_resource_data_sync.example example-name
```
