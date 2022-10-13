---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_delegated_admin_account"
description: |-
  Terraform resource for managing an AWS Inspector V2 Delegated Admin Account.
---

# Resource: aws_inspector2_delegated_admin_account

Terraform resource for managing an AWS Inspector V2 Delegated Admin Account.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_delegated_admin_account" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Delegated Admin Account. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.
* `example_attribute` - Concise description. Do not begin the description with "An", "The", "Defines", "Indicates", or "Specifies," as these are verbose. In other words, "Indicates the amount of storage," can be rewritten as "Amount of storage," without losing any information.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Inspector V2 Delegated Admin Account can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_inspector2_delegated_admin_account.example rft-8012925589
```
