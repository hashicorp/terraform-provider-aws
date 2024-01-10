---
subcategory: "Mainframe Modernization"
layout: "aws"
page_title: "AWS: aws_m2_environment"
description: |-
  Terraform resource for managing an AWS Mainframe Modernization Environment.
---
# Resource: aws_m2_environment

Terraform resource for managing an [AWS Mainframe Modernization Environment](https://docs.aws.amazon.com/m2/latest/userguide/environments-m2.html).

## Example Usage

### Basic Usage

```terraform
resource "aws_m2_environment" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Environment. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Mainframe Modernization Environment using the `example_id_arg`. For example:

```terraform
import {
  to = aws_m2_environment.example
  id = "environment-id-12345678"
}
```

Using `terraform import`, import Mainframe Modernization Environment using the `example_id_arg`. For example:

```console
% terraform import aws_m2_environment.example environment-id-12345678
```
