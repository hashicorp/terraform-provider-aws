---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system"
description: |-
  Lists S3 Files File System resources.
---

# List Resource: aws_s3files_file_system

Lists S3 Files File System resources.

## Example Usage

```terraform
list "aws_s3files_file_system" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `account` - (Optional) AWS account ID to query. Defaults to the account of the provider's credentials. See the [Per-Resource Account guide](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/guides/per-resource-account).
* `region` - (Optional) Region to query. Defaults to provider region.
