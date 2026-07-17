---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_buckets"
description: |-
  Provides details about AWS S3 (Simple Storage) buckets with optional filters.
---

# Data Source: aws_s3_buckets

Provides details about AWS S3 (Simple Storage) buckets with optional filters.

## Example Usage

```terraform
data "aws_s3_buckets" "example" {
  max_buckets = 3
  prefix      = "tf-"
  region      = "us-west-2"
}
```

### Basic Usage

```terraform
data "aws_s3_buckets" "example" {
}
```

## Argument Reference

The following arguments are optional:

* `max_buckets` - (Optional) Maximum number of buckets to return.
  * Valid Range: Minimum value of 1. Maximum value of 10000.
* `prefix` - (Optional) Limits the response to bucket names that begin with the specified bucket name prefix.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `buckets` - List of bucket objects:
   * `bucket_arn` - Bucket ARN.
   * `bucket_region` - Bucket region.
   * `creation_date` - Bucket creation date.
   * `name` - Bucket name.
