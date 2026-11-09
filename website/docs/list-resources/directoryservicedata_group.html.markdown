---
subcategory: "Directory Service Data"
layout: "aws"
page_title: "AWS: aws_directoryservicedata_group"
description: |-
  Lists Directory Service Data Group resources.
---

# List Resource: aws_directoryservicedata_group

Lists Directory Service Data Group resources.

## Example Usage

```terraform
list "aws_directoryservicedata_group" "example" {
  provider = aws

  config {
    directory_id = aws_directory_service_directory.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `directory_id` - (Required) ID of the Directory Service directory to list groups from.
* `region` - (Optional) Region to query. Defaults to provider region.
