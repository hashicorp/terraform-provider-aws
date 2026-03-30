---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_bucket_object_lock_configuration"
description: |-
  Provides details about an AWS S3 (Simple Storage) Bucket Object Lock Configuration.
---

# Data Source: aws_s3_bucket_object_lock_configuration

Provides details about an AWS S3 (Simple Storage) Bucket Object Lock Configuration.

## Example Usage

### Basic Usage

```terraform
data "aws_s3_bucket_object_lock_configuration" "example" {
  bucket = "example-bucket"
}
```

## Argument Reference

The following arguments are required:

* `bucket` - (Required) Name of the bucket.

The following arguments are optional:

* `expected_bucket_owner` - (Optional) Account ID of the expected bucket owner.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `object_lock_enabled` - Indicates whether this bucket has an Object Lock configuration enabled.
* `rule` - Object lock rule for the specified object. See [Rule](#rule) below.

### Rule

The `rule` block supports the following:

* `default_retention` - Default object lock retention settings for new objects placed in the bucket. See [Default Retention](#default-retention) below.

### Default Retention

The `default_retention` block supports the following:

* `days` - Default retention period in days.
* `mode` - Default object lock retention mode. Valid values are `GOVERNANCE` and `COMPLIANCE`.
* `years` - Default retention period in years.
