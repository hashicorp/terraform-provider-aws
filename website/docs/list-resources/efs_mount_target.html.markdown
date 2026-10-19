---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_mount_target"
description: |-
  Lists EFS Mount Target resources.
---

# List Resource: aws_efs_mount_target

Lists EFS Mount Target resources.

## Example Usage

```terraform
list "aws_efs_mount_target" "example" {
  provider = aws

  config {
    file_system_id = aws_efs_file_system.example.id
  }
}
```

## Argument Reference

This list resource supports the following arguments:

~> **Note:** One of `access_point_id` or `file_system_id` must be configured.

* `access_point_id` - (Optional) ID of the access point whose mount targets that you want to list.
* `file_system_id` - (Optional) ID of the file system whose mount targets you want to list.
* `region` - (Optional) Region to query. Defaults to provider region.
