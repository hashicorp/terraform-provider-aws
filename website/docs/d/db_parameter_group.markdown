---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_db_parameter_group"
description: |-
  Information about a database parameter group.
---

# Data Source: aws_db_parameter_group

Information about a database parameter group.

## Example Usage

```terraform
data "aws_db_parameter_group" "test" {
  name = "default.postgres15"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) DB parameter group name.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the parameter group.
* `family` - Family of the parameter group.
* `description` - Description of the parameter group.
* `parameter` - The DB parameters block. See [`parameter` Block](#parameter-block) below for more details.

### `parameter` Block

The `parameter` blocks support the following arguments:

* `name` - The name of the DB parameter.
* `value` - The value of the DB parameter.
* `apply_method` - "immediate" or "pending-reboot".
