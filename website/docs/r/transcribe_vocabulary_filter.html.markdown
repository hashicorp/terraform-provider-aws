---
subcategory: "Transcribe"
layout: "aws"
page_title: "AWS: aws_transcribe_vocabulary_filter"
description: |-
  Terraform resource for managing an AWS Transcribe VocabularyFilter.
---

# Resource: aws_transcribe_vocabulary_filter

Terraform resource for managing an AWS Transcribe VocabularyFilter.

## Example Usage

### Basic Usage

```terraform
resource "aws_transcribe_vocabulary_filter" "example" {
  vocabulary_filter_name = "example"
  language_code          = "en-US"
  words                  = ["cars", "bucket"]

  tags = {
    tag1 = "value1"
    tag2 = "value3"
  }
}
```

## Argument Reference

The following arguments are required:

* `example_arg` - (Required) Concise argument description.

The following arguments are optional:

* `optional_arg` - (Optional) Concise argument description.

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `id` - VocabularyFilter name.
* `arn` - ARN of the VocabularyFilter.
* `example_attribute` - Concise description.

## Import

Transcribe VocabularyFilter can be imported using the `vocabulary_filter_name`, e.g.,

```
$ terraform import aws_transcribe_vocabulary_filter.example example-name
```
