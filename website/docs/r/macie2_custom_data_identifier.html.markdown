---
subcategory: "Macie"
layout: "aws"
page_title: "AWS: aws_macie2_custom_data_identifier"
description: |-
  Provides a resource to manage an AWS Macie Custom Data Identifier.
---

# Resource: aws_macie2_custom_data_identifier

Provides a resource to manage an [AWS Macie Custom Data Identifier](https://docs.aws.amazon.com/macie/latest/APIReference/custom-data-identifiers-id.html).

## Example Usage

```terraform
resource "aws_macie2_account" "example" {}

resource "aws_macie2_custom_data_identifier" "example" {
  name                   = "NAME OF CUSTOM DATA IDENTIFIER"
  regex                  = "[0-9]{3}-[0-9]{2}-[0-9]{4}"
  description            = "DESCRIPTION"
  maximum_match_distance = 10
  keywords               = ["keyword"]
  ignore_words           = ["ignore"]

  depends_on = [aws_macie2_account.test]
}
```

## Argument Reference

This resource supports the following arguments:

* `description` - (Optional) Custom description of the custom data identifier. Can contain as many as 512 characters.
* `ignore_words` - (Optional) Array that lists specific character sequences (ignore words) to exclude from the results. If the text matched by the regular expression is the same as any string in this array, Amazon Macie ignores it. The array can contain as many as 10 ignore words. Each ignore word can contain 4 - 90 characters. Ignore words are case sensitive.
* `keywords` - (Optional) Array that lists specific character sequences (keywords), one of which must be within proximity (`maximum_match_distance`) of the regular expression to match. The array can contain as many as 50 keywords. Each keyword can contain 3 - 90 characters. Keywords aren't case sensitive.
* `maximum_match_distance` - (Optional) Maximum number of characters that can exist between text that matches the regex pattern and the character sequences specified by the keywords array. Macie includes or excludes a result based on the proximity of a keyword to text that matches the regex pattern. The distance can be 1 - 300 characters. The default value is 50.
* `name` - (Optional) Custom name for the custom data identifier. Can contain as many as 128 characters. If omitted, Terraform assigns a random, unique name. Conflicts with `name_prefix`.
* `name_prefix` - (Optional) Creates a unique name beginning with the specified prefix. Conflicts with `name`.
* `regex` - (Optional) Regular expression (regex) that defines the pattern to match. The expression can contain as many as 512 characters.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the custom data identifier.
* `created_at` - Date and time, in UTC and extended RFC 3339 format, when the custom data identifier was created.
* `id` - Unique identifier (ID) of the Macie custom data identifier.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `4m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import `aws_macie2_custom_data_identifier` using the ID. For example:

```terraform
import {
  to = aws_macie2_custom_data_identifier.example
  id = "abcd1"
}
```

Using `terraform import`, import `aws_macie2_custom_data_identifier` using the ID. For example:

```console
% terraform import aws_macie2_custom_data_identifier.example abcd1
```
