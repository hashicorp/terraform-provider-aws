---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_vocabulary"
description: |-
  Terraform resource for managing an AWS Transcribe Vocabulary.
---

# Resource: aws_transcribe_vocabulary

Terraform resource for managing an AWS Transcribe Vocabulary.

## Example Usage

### Basic Usage

```terraform
resource "aws_transcribe_vocabulary" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the Vocabulary.
* `example_attribute` - Concise description.

## Timeouts

`aws_transcribe_vocabulary` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

Transcribe Vocabulary can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_transcribe_vocabulary.example rft-8012925589
```
