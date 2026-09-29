---
subcategory: "Directory Service"
layout: "aws"
page_title: "AWS: aws_directory_service_directory_settings"
description: |-
  Lists Directory Service Directory Settings resources.
---

# List Resource: aws_directory_service_directory_settings

Lists Directory Service Directory Settings resources.

## Example Usage

```terraform
list "aws_directory_service_directory_settings" "example" {
  provider = aws

  directory_id = "d-1234567890"
}
```

## Argument Reference

This list resource supports the following arguments:

* `directory_id` - (Required) ID of the directory.
* `region` - (Optional) Region to query. Defaults to provider region.
