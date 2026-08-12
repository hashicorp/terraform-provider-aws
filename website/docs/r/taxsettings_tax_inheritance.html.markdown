---
subcategory: "Tax Settings"
layout: "aws"
page_title: "AWS: aws_taxsettings_tax_inheritance"
description: |-
  Manages tax information inheritance of the payer account within an AWS Organizations.
---

# Resource: aws_taxsettings_tax_inheritance

Manages tax information inheritance of the management account within an AWS Organizations.

~> **Note:** This resource requires the Organizations management account. The terms "payer account" and "management account" are used interchangeably, as AWS refers to the [managemement account as the payer account](https://docs.aws.amazon.com/organizations/latest/userguide/orgs_getting-started_concepts.html).

~> **Note:** The tax inheritance can only be updated once every 15 minutes. Attempting to update it more frequently triggers a `ConflictException`.

## Example Usage

### Basic Usage

```terraform
resource "aws_taxsettings_tax_inheritance" "example" {
  heritage_status = "OptIn"
}
```

## Argument Reference

The following arguments are required:

* `heritage_status` - (Required) Whether to enable (`OptIn`) or disable (`OptOut`) the tax information inheritance of the payer account for all member accounts within your AWS Organizations.

## Attribute Reference

This resource exports no additional attributes.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `10m`)
* `update` - (Default `10m`)

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_taxsettings_tax_inheritance.example
  identity = {}
}

resource "aws_taxsettings_tax_inheritance" "example" {
  heritage_status = "OptIn"
}
```

### Identity Schema

#### Optional

* `account_id` (String) AWS Account where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Tax Settings Tax Inheritance using the `example_id_arg`. For example:

```terraform
import {
  to = aws_taxsettings_tax_inheritance.example
  id = "123456789012"
}
```

Using `terraform import`, import Tax Settings Tax Inheritance using the `example_id_arg`. For example:

```console
% terraform import aws_taxsettings_tax_inheritance.example "123456789012"
```
