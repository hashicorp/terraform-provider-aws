---
subcategory: "DynamoDB Accelerator (DAX)"
layout: "aws"
page_title: "AWS: aws_dax_subnet_group"
description: |-
  Provides an DAX Subnet Group resource.
---

# Resource: aws_dax_subnet_group

Provides a DAX Subnet Group resource.

## Example Usage

```terraform
resource "aws_dax_subnet_group" "example" {
  name       = "example"
  subnet_ids = [aws_subnet.example1.id, aws_subnet.example2.id]
}
```

## Argument Reference

This resource supports the following arguments:

* `name` – (Required) The name of the subnet group.
* `description` - (Optional) A description of the subnet group.
* `subnet_ids` – (Required) A list of VPC subnet IDs for the subnet group.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - The name of the subnet group.
* `vpc_id` – VPC ID of the subnet group.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DAX Subnet Group using the `name`. For example:

```terraform
import {
  to = aws_dax_subnet_group.example
  id = "my_dax_sg"
}
```

Using `terraform import`, import DAX Subnet Group using the `name`. For example:

```console
% terraform import aws_dax_subnet_group.example my_dax_sg
```
