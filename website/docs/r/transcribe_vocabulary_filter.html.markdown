---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_vocabularyfilter"
description: |-
  Terraform resource for managing an AWS Transcribe VocabularyFilter.
---

# Resource: aws_transcribe_vocabularyfilter

Terraform resource for managing an AWS Transcribe VocabularyFilter.

## Example Usage

### Basic Usage

```terraform
resource "aws_transcribe_vocabularyfilter" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the VocabularyFilter.
* `example_attribute` - Concise description.

## Timeouts

`aws_transcribe_vocabularyfilter` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

Transcribe VocabularyFilter can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_transcribe_vocabularyfilter.example rft-8012925589
```
