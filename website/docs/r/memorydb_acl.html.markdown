---
subcategory: "MemoryDB"
layout: "aws"
page_title: "AWS: aws_memorydb_acl"
description: |-
  Provides a MemoryDB ACL.
---

# Resource: aws_memorydb_acl

Provides a MemoryDB ACL.

More information about users and ACL-s can be found in the [MemoryDB User Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/clusters.acls.html).

## Example Usage

```terraform
resource "aws_memorydb_acl" "example" {
  name       = "my-acl"
  user_names = ["my-user-1", "my-user-2"]
}
```

## Argument Reference

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the ACL. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `user_names` - (Optional) Set of MemoryDB user names to be included in this ACL.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the ACL.
* `id` - Same as `name`.
* `minimum_engine_version` - Minimum engine version supported by the ACL.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import an ACL using the `name`. For example:

```terraform
import {
  to = aws_memorydb_acl.example
  id = "my-acl"
}
```

Using `terraform import`, import an ACL using the `name`. For example:

```console
% terraform import aws_memorydb_acl.example my-acl
```
