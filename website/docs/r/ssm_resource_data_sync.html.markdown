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
resource "aws_ssm_resource_data_sync" "example" {
  name = "example"

  s3_destination {
    bucket_name = aws_s3_bucket.example.bucket
    region      = aws_s3_bucket.example.region
  }
}

resource "aws_s3_bucket" "example" {
  bucket = "example"
}

data "aws_iam_policy_document" "example" {
  statement {
    sid    = "SSMBucketPermissionsCheck"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ssm.amazonaws.com"]
    }

    actions   = ["s3:GetBucketAcl"]
    resources = [aws_s3_bucket.example.arn]
  }

  statement {
    sid    = "SSMBucketDelivery"
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["ssm.amazonaws.com"]
    }

    actions   = ["s3:PutObject"]
    resources = ["${aws_s3_bucket.example.arn}/*"]

    condition {
      test     = "StringEquals"
      variable = "s3:x-amz-acl"
      values   = ["bucket-owner-full-control"]
    }
  }
}

resource "aws_s3_bucket_policy" "example" {
  bucket = aws_s3_bucket.example.bucket
  policy = data.aws_iam_policy_document.example.json
}
```

## Argument Reference

This resource supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `name` - (Required) Name for the configuration.
* `s3_destination` - (Required) Amazon S3 configuration details for the sync.

### s3_destination

`s3_destination` supports the following:

* `bucket_name` - (Required) Name of S3 bucket where the aggregated data is stored.
* `destination_data_sharing` - (Optional) Enables destination data sharing.
  See [`destination_data_sharing` below](#destination_data_sharing).
* `region` - (Required) Region with the bucket targeted by the Resource Data Sync.
* `kms_key_arn` - (Optional) ARN of an encryption key for a destination in Amazon S3.
* `prefix` - (Optional) Prefix for the bucket.
* `sync_format` - (Optional) A supported sync format. Only JsonSerDe is currently supported. Defaults to JsonSerDe.

### destination_data_sharing

`destination_data_sharing` supports the following:

* `destination_data_sharing_type` - (Optional) Data sharing type.
  Only `Organization` is supported.

## Attribute Reference

This resource exports no additional attributes.

**Note:** If `s3_destination.destination_data_sharing` is set, the imported resource will be replaced on the next `terrafrom apply`.

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
