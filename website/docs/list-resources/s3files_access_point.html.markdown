---
subcategory: "S3 Files"
layout: "aws"
page_title: "AWS: aws_s3files_access_point"
description: |-
  Lists S3 Files Access Point resources.
---

# List Resource: aws_s3files_access_point

Lists S3 Files Access Point resources.

## Example Usage

```terraform
list "aws_s3files_access_point" "example" {
  provider = aws
}
```

## Argument Reference

This list resource supports the following arguments:

* `region` - (Optional) Region to query. Defaults to provider region.
