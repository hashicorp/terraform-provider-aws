---
subcategory: "MediaLive"
layout: "aws"
page_title: "AWS: aws_medialive_input"
description: |-
  Terraform resource for managing an AWS MediaLive Input.
---

# Resource: aws_medialive_input

Terraform resource for managing an AWS MediaLive Input.

## Example Usage

### Basic Usage

```terraform
resource "aws_medialive_input" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Input.
* `example_attribute` - Concise description.

## Timeouts

[Configuration options](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

MediaLive Input can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_medialive_input.example rft-8012925589
```
