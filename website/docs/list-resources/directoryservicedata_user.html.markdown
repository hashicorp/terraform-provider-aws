---
subcategory: "Directory Service Data"
layout: "aws"
page_title: "AWS: aws_directoryservicedata_user"
description: |-
  Lists Directory Service Data User resources.
---

# List Resource: aws_directoryservicedata_user

Lists Directory Service Data User resources.

## Example Usage

```terraform
list "aws_directoryservicedata_user" "example" {
  provider = aws

  config {
    directory_id = aws_directory_service_directory.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

* `directory_id` - (Required) ID of the Directory Service directory to list users from.
* `region` - (Optional) Region to query. Defaults to provider region.
