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

* `region` - (Optional) Region to query. Defaults to provider region.
