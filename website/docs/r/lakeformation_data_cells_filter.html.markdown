---
subcategory: "Lake Formation"
layout: "aws"
page_title: "AWS: aws_lakeformation_data_cells_filter"
description: |-
  Terraform resource for managing an AWS Lake Formation Data Cells Filter.
---
# Resource: aws_lakeformation_data_cells_filter

Terraform resource for managing an AWS Lake Formation Data Cells Filter.

## Example Usage

### Basic Usage

```terraform
resource "aws_lakeformation_data_cells_filter" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Data Cells Filter. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lake Formation Data Cells Filter using the `example_id_arg`. For example:

```terraform
import {
  to = aws_lakeformation_data_cells_filter.example
  id = "data_cells_filter-id-12345678"
}
```

Using `terraform import`, import Lake Formation Data Cells Filter using the `example_id_arg`. For example:

```console
% terraform import aws_lakeformation_data_cells_filter.example data_cells_filter-id-12345678
```
