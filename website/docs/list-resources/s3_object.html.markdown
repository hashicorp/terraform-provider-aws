---
subcategory: "S3 (Simple Storage)"
layout: "aws"
page_title: "AWS: aws_s3_object"
description: |-
  Lists S3 (Simple Storage) Object resources.
---

# List Resource: aws_s3_object

Lists S3 (Simple Storage) Object resources.

## Example Usage

```terraform
list "aws_s3_object" "example" {
  provider = aws
  config {
    bucket = "example"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `bucket` - (Required) Name of the S3 bucket to list objects from.
* `prefix` - (Optional) Limits the response to keys that begin with the specified prefix.
* `region` - (Optional) Region to query. Defaults to provider region.
