---
subcategory: "Polly"
layout: "aws"
page_title: "AWS: aws_polly_voices"
description: |-
  Terraform data source for managing an AWS Polly Voices.
---

# Data Source: aws_polly_voices

Terraform data source for managing an AWS Polly Voices.

## Example Usage

### Basic Usage

```terraform
data "aws_polly_voices" "example" {}
```

### With Language Code

```terraform
data "aws_polly_voices" "example" {
  language_code = "en-GB"
}
```

## Argument Reference

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `engine` - (Optional) Engine used by Amazon Polly when processing input text for speech synthesis. Valid values are `standard`, `neural`, and `long-form`.
* `include_additional_language_codes` - (Optional) Whether to return any bilingual voices that use the specified language as an additional language.
* `language_code` - (Optional) Language identification tag for filtering the list of voices returned. If not specified, all available voices are returned.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `id` - AWS account ID.
* `voices` - List of voices with their properties. See [`voices` Attribute Reference](#voices-attribute-reference) below.

### `voices` Attribute Reference

See the [AWS Polly Voice documentation](https://docs.aws.amazon.com/polly/latest/dg/API_Voice.html) for additional details.

* `additional_language_codes` - Additional codes for languages available for the specified voice in addition to its default language.
* `gender` - Gender of the voice.
* `id` - Amazon Polly assigned voice ID.
* `language_code` - Language code of the voice.
* `language_name` - Human readable name of the language in English.
* `name` - Name of the voice.
* `supported_engines` - Specifies which engines are supported by a given voice.
