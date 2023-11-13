---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_cluster_parameter_group"
description: |-
  Manages a Neptune Cluster Parameter Group
---

# Resource: aws_neptune_cluster_parameter_group

Manages a Neptune Cluster Parameter Group

## Example Usage

```terraform
resource "aws_neptune_cluster_parameter_group" "example" {
  family      = "neptune1"
  name        = "example"
  description = "neptune cluster parameter group"

  parameter {
    name  = "neptune_enable_audit_log"
    value = 1
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Optional, Forces new resource) The name of the neptune cluster parameter group. If omitted, Terraform will assign a random, unique name.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `family` - (Required) The family of the neptune cluster parameter group.
* `description` - (Optional) The description of the neptune cluster parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of neptune parameters to apply.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Parameter blocks support the following:

* `name` - (Required) The name of the neptune parameter.
* `value` - (Required) The value of the neptune parameter.
* `apply_method` - (Optional) Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The neptune cluster parameter group name.
* `arn` - The ARN of the neptune cluster parameter group.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Neptune Cluster Parameter Groups using the `name`. For example:

```terraform
import {
  to = aws_neptune_cluster_parameter_group.cluster_pg
  id = "production-pg-1"
}
```

Using `terraform import`, import Neptune Cluster Parameter Groups using the `name`. For example:

```console
% terraform import aws_neptune_cluster_parameter_group.cluster_pg production-pg-1
```
