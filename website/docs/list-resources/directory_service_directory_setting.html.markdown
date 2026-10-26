---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_directory_setting"
description: |-
  Lists Directory Service Directory Setting resources.
---

# List Resource: aws_directory_service_directory_setting

Lists Directory Service Directory Setting resources.

## Example Usage

```terraform
list "aws_directory_service_directory_setting" "example" {
  provider = aws

  config {
    directory_id = "d-1234567890"
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `directory_id` - (Required) ID of the directory.
* `region` - (Optional) Region to query. Defaults to provider region.
