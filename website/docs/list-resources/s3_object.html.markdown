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
  bucket   = "my-bucket-name"
}
```

## Argument Reference

This list resource supports the following arguments:

* `bucket` - (Required) Name of the S3 bucket to list objects from.
* `region` - (Optional) Region to query. Defaults to provider region.
