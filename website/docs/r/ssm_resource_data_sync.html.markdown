---
layout: "aws"
page_title: "AWS: aws_ssm_resource_data_sync"
sidebar_current: "docs-aws-resource-ssm-resource-data-sync"
description: |-
  Provides a SSM resource data sync.
---

# aws_athena_database

Provides a SSM resource data sync.

## Example Usage

```hcl
resource "aws_s3_bucket" "hoge" {
  bucket = "tf-test-bucket-1234"
  region = "us-east-1"
}

resource "aws_s3_bucket_policy" "hoge" {
  bucket = "${aws_s3_bucket.hoge.bucket}"
  policy = <<EOF
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "SSMBucketPermissionsCheck",
            "Effect": "Allow",
            "Principal": {
                "Service": "ssm.amazonaws.com"
            },
            "Action": "s3:GetBucketAcl",
            "Resource": "arn:aws:s3:::tf-test-bucket-1234"
        },
        {
            "Sid": " SSMBucketDelivery",
            "Effect": "Allow",
            "Principal": {
                "Service": "ssm.amazonaws.com"
            },
            "Action": "s3:PutObject",
            "Resource": ["arn:aws:s3:::tf-test-bucket-1234/*"],
            "Condition": {
                "StringEquals": {
                    "s3:x-amz-acl": "bucket-owner-full-control"
                }
            }
        }
      ]
  }
  EOF
}

resource "aws_ssm_resource_data_sync" "foo" {
  name = "foo"
  destination = {
    bucket = "${aws_s3_bucket.hoge.bucket}"
    region = "${aws_s3_bucket.hoge.region}"
  }
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required) Name for the configuration.
* `destination` - (Required) Amazon S3 configuration details for the sync.

## destination

`destination` supports the following:

* `bucket` - (Required) Name of S3 bucket where the aggregated data is stored.
* `region` - (Required) Region with the bucket targeted by the Resource Data Sync.
* `key` - (Optional) ARN of an encryption key for a destination in Amazon S3.
* `prefix` - (Optional) Prefix for the bucket.
