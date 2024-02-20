---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_bot_version"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Bot Version.
---

# Resource: aws_lexv2models_bot_version

Terraform resource for managing an AWS Lex V2 Models Bot Version.

## Example Usage

### Basic Usage

```terraform
resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models.test.id
  locale_specification = {
    "en_US" = {
      source_bot_version = "DRAFT"
    }
  }
}
```

## Argument Reference

This resource supports the following arguments:

* `bot_id` - (Required) Idientifier of the bot to create the version for.
* `locale_specification` - (Required) Specifies the locales that Amazon Lex adds to this version. You can choose the draft version or any other previously published version for each locale. When you specify a source version, the locale data is copied from the source version to the new version.

   The attribute value is a map with one or more entries, each of which has a locale name as the key and an object with the following attribute as the value:
    * `sourceBotVersion` - (Required) The version of a bot used for a bot locale. Valid values: `DRAFT`, a numeric version.
* `description` - (Optional) A description of the version. Use the description to help identify the version in lists.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `bot_version` - Version number assigned to the version.
* `id` - A comma-delimited string concatinating `bot_id` and `bot_version`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Bot Version using the `id`. For example:

```terraform
import {
  to = aws_lexv2models_bot_version.example
  id = "id-12345678,1"
}
```

Using `terraform import`, import Lex V2 Models Bot Version using the `id`. For example:

```console
% terraform import aws_lexv2models_bot_version.example id-12345678,1
```
