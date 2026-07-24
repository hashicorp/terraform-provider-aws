---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_synchronization_configuration"
description: |-
  Lists S3 Files Synchronization resources.
---

# List Resource: aws_s3files_synchronization_configuration

Lists S3 Files Synchronization resources.

## Example Usage

```terraform
list "aws_s3files_synchronization_configuration" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
