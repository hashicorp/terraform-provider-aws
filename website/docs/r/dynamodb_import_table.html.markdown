---
subcategory: "DynamoDB"
layout: "aws"
page_title: "AWS: aws_dynamodb_import_table"
description: |-
  Terraform resource for managing an AWS DynamoDB Import Table.
---
# Resource: aws_dynamodb_import_table

Terraform resource for managing an AWS DynamoDB Import Table.

## Example Usage

### Basic Usage

```terraform
resource "aws_dynamodb_import_table" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Import Table. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import DynamoDB Import Table using the `example_id_arg`. For example:

```terraform
import {
  to = aws_dynamodb_import_table.example
  id = "import_table-id-12345678"
}
```

Using `terraform import`, import DynamoDB Import Table using the `example_id_arg`. For example:

```console
% terraform import aws_dynamodb_import_table.example import_table-id-12345678
```
