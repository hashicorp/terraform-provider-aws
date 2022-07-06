---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_medicalvocabulary"
description: |-
  Terraform resource for managing an AWS Transcribe MedicalVocabulary.
---

# Resource: aws_transcribe_medicalvocabulary

Terraform resource for managing an AWS Transcribe MedicalVocabulary.

## Example Usage

### Basic Usage

```terraform
resource "aws_transcribe_medicalvocabulary" "example" {
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - ARN of the MedicalVocabulary.
* `example_attribute` - Concise description.

## Timeouts

`aws_transcribe_medicalvocabulary` provides the following [Timeouts](https://www.terraform.io/docs/configuration/blocks/resources/syntax.html#operation-timeouts) configuration options:

* `create` - (Optional, Default: `60m`)
* `update` - (Optional, Default: `180m`)
* `delete` - (Optional, Default: `90m`)

## Import

Transcribe MedicalVocabulary can be imported using the `example_id_arg`, e.g.,

```
$ terraform import aws_transcribe_medicalvocabulary.example rft-8012925589
```
