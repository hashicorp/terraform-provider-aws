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

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Bucket ARNs.
* `buckets` - Buckets names.
