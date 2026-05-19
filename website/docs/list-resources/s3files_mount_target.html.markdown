---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_mount_target"
description: |-
  Lists S3 Files Mount Target resources.
---

# List Resource: aws_s3files_mount_target

Lists S3 Files Mount Target resources.

## Example Usage

```terraform
list "aws_s3files_mount_target" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
