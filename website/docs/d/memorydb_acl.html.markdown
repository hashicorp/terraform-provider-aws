---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_acl"
description: |-
  Provides information about a MemoryDB ACL.
---

# Resource: aws_memorydb_acl

Provides information about a MemoryDB ACL.

## Example Usage

```terraform
data "aws_memorydb_acl" "example" {
  name = "my-acl"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the ACL.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - Name of the ACL.
* `arn` - ARN of the ACL.
* `minimum_engine_version` - The minimum engine version supported by the ACL.
* `tags` - Map of tags assigned to the ACL.
* `user_names` - Set of MemoryDB user names included in this ACL.
