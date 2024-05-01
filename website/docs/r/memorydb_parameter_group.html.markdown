---
subcategory: "MemoryDB for Redis"
layout: "aws"
page_title: "AWS: aws_memorydb_parameter_group"
description: |-
  Provides a MemoryDB Parameter Group.
---

# Resource: aws_memorydb_parameter_group

Provides a MemoryDB Parameter Group.

More information about parameter groups can be found in the [MemoryDB User Guide](https://docs.aws.amazon.com/memorydb/latest/devguide/parametergroups.html).

## Example Usage

```terraform
resource "aws_memorydb_parameter_group" "example" {
  name   = "my-parameter-group"
  family = "memorydb_redis6"

  parameter {
    name  = "activedefrag"
    value = "yes"
  }
}
```

## Argument Reference

The following arguments are required:

* `family` - (Required, Forces new resource) The engine version that the parameter group can be used with.

The following arguments are optional:

* `name` - (Optional, Forces new resource) Name of the parameter group. If omitted, Terraform will assign a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `description` - (Optional, Forces new resource) Description for the parameter group. Defaults to `"Managed by Terraform"`.
* `parameter` - (Optional) Set of MemoryDB parameters to apply. Any parameters not specified will fall back to their family defaults. Detailed below.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### parameter Configuration Block

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value of the parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Same as `name`.
* `arn` - The ARN of the parameter group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import a parameter group using the `name`. For example:

```terraform
import {
  to = aws_memorydb_parameter_group.example
  id = "my-parameter-group"
}
```

Using `terraform import`, import a parameter group using the `name`. For example:

```console
% terraform import aws_memorydb_parameter_group.example my-parameter-group
```
