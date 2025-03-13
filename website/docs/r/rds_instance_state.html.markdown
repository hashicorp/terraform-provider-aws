---
subcategory: "RDS (Relational Database)"
layout: "aws"
page_title: "AWS: aws_rds_instance_state"
description: |-
  Terraform resource for managing an AWS RDS (Relational Database) RDS Instance State.
---

# Resource: aws_rds_instance_state

Terraform resource for managing an AWS RDS (Relational Database) RDS Instance State.

~> Destruction of this resource is a no-op and **will not** modify the instance state

## Example Usage

### Basic Usage

```terraform
resource "aws_rds_instance_state" "test" {
  identifier = aws_db_instance.test.identifier
  state      = "available"
}
```

## Argument Reference

The following arguments are required:

* `identifier` - (Required) DB Instance Identifier
* `state` - (Required) Configured state of the DB Instance. Valid values are `available` and `stopped`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `identifier` - DB Instance Identifier

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import RDS (Relational Database) RDS Instance State using the `example_id_arg`. For example:

```terraform
import {
  to = aws_rds_instance_state.example
  id = "db-L72FUFBZX2RRXT3HOJSIUQVOKE"
}
```

Using `terraform import`, import RDS (Relational Database) RDS Instance State using the `example_id_arg`. For example:

```console
% terraform import aws_rds_instance_state.example rds_instance_state-id-12345678
```
