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
  locale_id   = aws_lexv2models_locale.example.locale_id
  name        = "example"
}
```

## Argument Reference

The following arguments are required:

* `bot_id` - (Required)
* `bot_version` - (Required)
* `intent_id` - (Required)
* `locale_id` - (Required)
* `name` - (Required)
* `value_elicitation_setting` - (Required)

The following arguments are optional:

* `description` - (Optional)
* `multiple_values_setting` - (Optional)
* `obfuscation_setting` - (Optional)
* `slot_type_id` - (Optional)
* `sub_slot_setting` - (Optional)

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
  id = "slot-id-12345678"
}
```

Using `terraform import`, import Lex V2 Models Slot using the `id`. For example:

```console
% terraform import aws_lexv2models_slot.example slot-id-12345678
```
