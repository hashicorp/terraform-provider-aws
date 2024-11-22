---
subcategory: "DynamoDB Accelerator (DAX)"
layout: "aws"
page_title: "AWS: aws_dax_parameter_group"
description: |-
  Provides an DAX Parameter Group resource.
---

# Resource: aws_dax_parameter_group

Provides a DAX Parameter Group resource.

## Example Usage

```terraform
resource "aws_dax_parameter_group" "example" {
  name = "example"

  parameters {
    name  = "query-ttl-millis"
    value = "100000"
  }

  parameters {
    name  = "record-ttl-millis"
    value = "100000"
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `name` – (Required) The name of the parameter group.

* `description` - (Optional, ForceNew) A description of the parameter group.

* `parameters` – (Optional) The parameters of the parameter group.

## parameters

`parameters` supports the following:

* `name` - (Required) The name of the parameter.
* `value` - (Required) The value for the parameter.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the parameter group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DAX Parameter Group using the `name`. For example:

```terraform
import {
  to = aws_dax_parameter_group.example
  id = "my_dax_pg"
}
```

Using `terraform import`, import DAX Parameter Group using the `name`. For example:

```console
% terraform import aws_dax_parameter_group.example my_dax_pg
```
