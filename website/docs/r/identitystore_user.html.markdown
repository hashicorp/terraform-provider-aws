---
subcategory: "IdentityStore"
layout: "aws"
page_title: "AWS: aws_identitystore_user"
description: |-
  Terraform resource for managing an AWS IdentityStore User.
---

# Resource: aws_identitystore_user

Terraform resource for managing an AWS IdentityStore User.

## Example Usage

### Basic Usage

```terraform
resource "aws_identitystore_user" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the User.
* `example_attribute` - Concise description.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

IdentityStore User can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_identitystore_user.example rft-8012925589
```
