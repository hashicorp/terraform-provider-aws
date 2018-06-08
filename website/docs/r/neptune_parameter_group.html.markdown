---
layout: "aws"
page_title: "AWS: aws_neptune_parameter_group"
sidebar_current: "docs-aws-resource-aws-neptune-parameter-group"
description: |-
   Provides an Neptune parameter group resource.
---

# aws_neptune_parameter_group

 Creates a parameter group for AWS Neptune

## Example Usage

```hcl
resource "aws_neptune_parameter_group" "bar" {
	name = "my_group"
	family = "neptune1"
	description = "Test parameter group for terraform"
}
```

```hcl
resource "aws_neptune_parameter_group" "bar" {
	name = "my_group"
	family = "neptune1"
	description = "Test parameter group for terraform"
	
	parameter {
	  name = "neptune_query_timeout"
      apply_method = "pending-reboot"
	  value = "25"
	}
}
```

## Argument Reference

The following arguments are supported:

* `name` - (Required, Forces new resource) The name of the Neptune parameter group.
* `family` - (Required) The family of the Neptune parameter group.
* `description` - (Optional) The description of the Neptune parameter group. Defaults to "Managed by Terraform".
* `parameter` - (Optional) A list of Neptune parameters to apply.
* `tags`  - (Optional) A mapping of tags to assign to the resource.

Parameter blocks support the following:

* `name`  - (Required) The name of the Neptune parameter.
* `value` - (Required) The value of the Neptune parameter.


## Attributes Reference

The following attributes are exported:

* `id` - The dbNeptune parameter group name.

## Import

Neptune Parameter groups can be imported using the `name`, e.g.

```
$ terraform import aws_neptune_parameter_group.some_pg some-pg
```
