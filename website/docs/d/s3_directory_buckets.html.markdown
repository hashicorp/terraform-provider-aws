---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_directory_buckets"
description: |-
  Lists Amazon S3 Express directory buckets.
---

# Data Source: aws_s3_directory_buckets

Lists Amazon S3 Express directory buckets.

## Example Usage

```terraform
data "aws_s3_directory_buckets" "example" {}
```

## Argument Reference

There are no arguments available for this data source.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Bucket ARNs.
* `buckets` - Buckets names.
