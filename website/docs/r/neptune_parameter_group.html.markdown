---
subcategory: "Neptune"
layout: "aws"
page_title: "AWS: aws_neptune_parameter_group"
description: |-
  Manages a Neptune Parameter Group
---

# Resource: aws_neptune_parameter_group

Manages a Neptune Parameter Group

## Example Usage

```terraform
resource "aws_neptune_parameter_group" "example" {
  family = "neptune1"
  name   = "example"

  parameter {
    name  = "neptune_query_timeout"
    value = "25"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` - (Required, Forces new resource) The name of the Neptune parameter group.
* `family` - (Required) The family of the Neptune parameter group.
* `description` - (Optional) The description of the Neptune parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of Neptune parameters to apply.
* `tags` - (Optional) A map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

Parameter blocks support the following:

* `name`  - (Required) The name of the Neptune parameter.
* `value` - (Required) The value of the Neptune parameter.
* `apply_method` - (Optional) The apply method of the Neptune parameter. Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The Neptune parameter group name.
* `arn` - The Neptune parameter group Amazon Resource Name (ARN).
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Neptune Parameter Groups using the `name`. For example:

```terraform
import {
  to = aws_neptune_parameter_group.some_pg
  id = "some-pg"
}
```

Using `terraform import`, import Neptune Parameter Groups using the `name`. For example:

```console
% terraform import aws_neptune_parameter_group.some_pg some-pg
```
