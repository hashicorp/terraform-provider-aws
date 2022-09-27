---
subcategory: "Inspector V2"
layout: "aws"
page_title: "AWS: aws_inspector2_organization_configuration"
description: |-
  Terraform resource for managing an AWS Inspector2 Organization Configuration.
---

# Resource: aws_inspector2_organization_configuration

Terraform resource for managing an AWS Inspector2 Organization Configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_inspector2_organization_configuration" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Organization Configuration.
* `example_attribute` - Concise description.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

Inspector2 Organization Configuration can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_inspector2_organization_configuration.example rft-8012925589
```
