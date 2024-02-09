---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_slot"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Slot.
---

# Resource: aws_lexv2models_slot

Terraform resource for managing an AWS Lex V2 Models Slot.

## Example Usage

### Basic Usage

```terraform
resource "aws_lexv2models_slot" "example" {
  bot_id      = aws_lexv2models_bot.example.id
  bot_version = aws_lexv2models_bot_version.example.bot_version
  intent_id   = aws_lexv2models_intent.example.id
  locale_id   = aws_lexv2models_bot_locale.example.locale_id
  name        = "example"
}
```

## Argument Reference

The following arguments are required:

* `bot_id` - (Required) Identifier of the bot associated with the slot.
* `bot_version` - (Required) Version of the bot associated with the slot.
* `intent_id` - (Required) Identifier of the intent that contains the slot.
* `locale_id` - (Required) Identifier of the language and locale that the slot will be used in.
* `name` - (Required) Name of the slot.
* `value_elicitation_setting` - (Required) Prompts that Amazon Lex sends to the user to elicit a response that provides the value for the slot.

The following arguments are optional:

* `description` - (Optional) Description of the slot.
* `multiple_values_setting` - (Optional) Whether the slot returns multiple values in one response. See the [`multiple_values_setting` argument reference](#multiple_values_setting-argument-reference) below.
* `obfuscation_setting` - (Optional) Determines how slot values are used in Amazon CloudWatch logs. See the [`obfuscation_setting` argument reference](#obfuscation_setting-argument-reference) below.
* `slot_type_id` - (Optional) Unique identifier for the slot type associated with this slot.
* `sub_slot_setting` - (Optional) Specifications for the constituent sub slots and the expression for the composite slot.

### `multiple_values_setting` Argument Reference

* `allow_multiple_values` - (Optional) Whether a slot can return multiple values. When `true`, the slot may return more than one value in a response. When `false`, the slot returns only a single value. Multi-value slots are only available in the `en-US` locale.

### `obfuscation_setting` Argument Reference

* `obfuscation_setting_type` - (Required) Whether Amazon Lex obscures slot values in conversation logs. Valid values are `DefaultObfuscation` and `None`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - A comma-delimited string concatenating `bot_id`, `bot_version`, `intent_id`, `locale_id`, and `slot_id`.
* `slot_id` - Unique identifier associated with the slot.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Slot using the `id`. For example:

```terraform
import {
  to = aws_lexv2models_slot.example
  id = "bot-1234,1,intent-5678,en-US,slot-9012"
}
```

Using `terraform import`, import Lex V2 Models Slot using the `id`. For example:

```console
% terraform import aws_lexv2models_slot.example bot-1234,1,intent-5678,en-US,slot-9012
```
