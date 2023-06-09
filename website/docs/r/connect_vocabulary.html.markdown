---
subcategory: "Connect"
layout: "aws"
page_title: "AWS: aws_connect_vocabulary"
description: |-
  Provides details about a specific Amazon Connect Vocabulary
---

# Resource: aws_connect_vocabulary

Provides an Amazon Connect Vocabulary resource. For more information see
[Amazon Connect: Getting Started](https://docs.aws.amazon.com/connect/latest/adminguide/amazon-connect-get-started.html)

## Example Usage

```terraform
resource "aws_connect_vocabulary" "example" {
  instance_id   = "aaaaaaaa-bbbb-cccc-dddd-111111111111"
  name          = "example"
  content       = "Phrase\tIPA\tSoundsLike\tDisplayAs\nLos-Angeles\t\t\tLos Angeles\nF.B.I.\tɛ f b i aɪ\t\tFBI\nEtienne\t\teh-tee-en\t"
  language_code = "en-US"
  tags = {
    "Key1" = "Value1"
  }
}
```

## Argument Reference

The following arguments are supported:

* `content` - (Required) The content of the custom vocabulary in plain-text format with a table of values. Each row in the table represents a word or a phrase, described with Phrase, IPA, SoundsLike, and DisplayAs fields. Separate the fields with TAB characters. For more information, see [Create a custom vocabulary using a table](https://docs.aws.amazon.com/transcribe/latest/dg/custom-vocabulary.html#create-vocabulary-table). Minimum length of `1`. Maximum length of `60000`.
* `instance_id` - (Required) Specifies the identifier of the hosting Amazon Connect Instance.
* `language_code` - (Required) The language code of the vocabulary entries. For a list of languages and their corresponding language codes, see [What is Amazon Transcribe?](https://docs.aws.amazon.com/transcribe/latest/dg/transcribe-whatis.html). Valid Values are `ar-AE`, `de-CH`, `de-DE`, `en-AB`, `en-AU`, `en-GB`, `en-IE`, `en-IN`, `en-US`, `en-WL`, `es-ES`, `es-US`, `fr-CA`, `fr-FR`, `hi-IN`, `it-IT`, `ja-JP`, `ko-KR`, `pt-BR`, `pt-PT`, `zh-CN`.
* `name` - (Required) A unique name of the custom vocabulary. Must not be more than 140 characters.
* `tags` - (Optional) Tags to apply to the vocabulary. If configured with a provider
[`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `delete` - (Default `100m`)

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `arn` - The Amazon Resource Name (ARN) of the vocabulary.
* `failure_reason` - The reason why the custom vocabulary was not created.
* `id` - The identifier of the hosting Amazon Connect Instance and identifier of the vocabulary
separated by a colon (`:`).
* `last_modified_time` - The timestamp when the custom vocabulary was last modified.
* `state` - The current state of the custom vocabulary. Valid values are `CREATION_IN_PROGRESS`, `ACTIVE`, `CREATION_FAILED`, `DELETE_IN_PROGRESS`.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `vocabulary_id` - The identifier of the custom vocabulary.

## Import

Amazon Connect Vocabularies can be imported using the `instance_id` and `vocabulary_id` separated by a colon (`:`), e.g.,

```
$ terraform import aws_connect_vocabulary.example f1288a1f-6193-445a-b47e-af739b2:c1d4e5f6-1b3c-1b3c-1b3c-c1d4e5f6c1d4e5
```
