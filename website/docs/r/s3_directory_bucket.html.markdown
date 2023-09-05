---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_directory_bucket"
description: |-
  Provides an Amazon S3 Express directory bucket resource.
---

# Resource: aws_s3_directory_bucket

Provides an Amazon S3 Express directory bucket resource.

## Example Usage

```terraform
resource "aws_s3_directory_bucket" "example" {
  bucket = "example--usw2-az2-d-s3"
}
```

## Argument Reference

This resource supports the following arguments:

* `bucket` - (Required) Name of the bucket.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Name of the bucket.
* `arn` - ARN of the bucket.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an Amazon S3 Express directory bucket using `bucket`. For example:

```terraform
import {
  to = aws_s3_directory_bucket.example
  id = "example--usw2-az2-d-s3"
}
```

Using `terraform import`, import S3 bucket using `bucket`. For example:

```console
% terraform import aws_s3_directory_bucket.example example--usw2-az2-d-s3
```
