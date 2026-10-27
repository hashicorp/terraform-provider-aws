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

* `language_code` - (Required) Language code you selected for your vocabulary filter. Refer to the [supported languages](https://docs.aws.amazon.com/transcribe/latest/dg/supported-languages.html) page for accepted codes.
* `vocabulary_filter_name` - (Required) Name of the VocabularyFilter.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the VocabularyFilter. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `vocabulary_filter_file_uri` - (Optional) Amazon S3 location (URI) of the text file that contains your custom VocabularyFilter. Conflicts with `words` argument.
* `words` - (Optional) List of terms to include in the vocabulary. Conflicts with `vocabulary_filter_file_uri` argument.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the VocabularyFilter.
* `download_uri` - Generated download URI.
* `id` - VocabularyFilter name.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Transcribe VocabularyFilter using the `vocabulary_filter_name`. For example:

```terraform
import {
  to = aws_transcribe_vocabulary_filter.example
  id = "example-name"
}
```

Using `terraform import`, import Transcribe VocabularyFilter using the `vocabulary_filter_name`. For example:

```console
% terraform import aws_transcribe_vocabulary_filter.example example-name
```
