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

* `description` - (Optional) Description of the Neptune parameter group. Defaults to "Managed by Terraform".
* `family` - (Required) Family of the Neptune parameter group.
* `name` - (Optional, Forces new resource) Name of the Neptune parameter group.
* `name_prefix` - (Optional, Forces new resource) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `parameter` - (Optional) List of Neptune parameters to apply.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `parameter` Block

The `parameter` block supports the following:

* `apply_method` - (Optional) Apply method of the Neptune parameter. Valid values are `immediate` and `pending-reboot`. Defaults to `pending-reboot`.
* `name` - (Required) Name of the Neptune parameter.
* `value` - (Required) Value of the Neptune parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - Neptune parameter group ARN.
* `id` - Neptune parameter group name.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

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
