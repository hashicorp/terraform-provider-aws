---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_file_system_policy"
description: |-
  Lists S3 Files File System Policy resources.
---

# List Resource: aws_s3files_file_system_policy

Lists S3 Files File System Policy resources.

## Example Usage

```terraform
list "aws_s3files_file_system_policy" "example" {
  provider       = aws
  file_system_id = "fs-12345678"
}
```

## Argument Reference

This list resource supports the following arguments:

* `file_system_id` - (Required) File system ID.
* `region` - (Optional) Region to query. Defaults to provider region.
