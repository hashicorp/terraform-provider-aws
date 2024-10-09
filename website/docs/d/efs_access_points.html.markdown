---
subcategory: "EFS (Elastic File System)"
layout: "aws"
page_title: "AWS: aws_efs_access_points"
description: |-
  Provides information about multiple Elastic File System (EFS) Access Points.
---

# Data Source: aws_efs_access_points

Provides information about multiple Elastic File System (EFS) Access Points.

## Example Usage

```terraform
data "aws_efs_access_points" "test" {
  file_system_id = "fs-12345678"
}
```

## Argument Reference

This data source supports the following arguments:

* `file_system_id` - (Required) EFS File System identifier.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arns` - Set of Amazon Resource Names (ARNs).
* `id` - EFS File System identifier.
* `ids` - Set of identifiers.
