---
subcategory: "Transit Gateway"
layout: "aws"
page_title: "AWS: aws_ec2_transit_gateway_default_route_table_association"
description: |-
  Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association.
---
# Resource: aws_ec2_transit_gateway_default_route_table_association

Terraform resource for managing an AWS EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association.

## Example Usage

### Basic Usage

```terraform
resource "aws_ec2_transit_gateway_default_route_table_association" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Transit Gateway Default Route Table Association. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association using the `example_id_arg`. For example:

```terraform
import {
  to = aws_ec2_transitgateway_default_route_table_association.example
  id = "transitgateway_default_route_table_association-id-12345678"
}
```

Using `terraform import`, import EC2 (Elastic Compute Cloud) Transit Gateway Default Route Table Association using the `example_id_arg`. For example:

```console
% terraform import aws_ec2_transitgateway_default_route_table_association.example transitgateway_default_route_table_association-id-12345678
```
