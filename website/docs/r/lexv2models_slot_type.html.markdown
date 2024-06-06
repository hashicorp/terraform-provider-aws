---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_slot_type"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Slot Type.
---

# Resource: aws_lexv2models_slot_type

Terraform resource for managing an AWS Lex V2 Models Slot Type.

## Example Usage

### Basic Usage

```terraform
resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonLexFullAccess"
}

resource "aws_lexv2models_bot" "test" {
  name                        = "testbot"
  idle_session_ttl_in_seconds = 60
  role_arn                    = aws_iam_role.test.arn

  data_privacy {
    child_directed = true
  }
}

resource "aws_lexv2models_bot_locale" "test" {
  locale_id                        = "en_US"
  bot_id                           = aws_lexv2models_bot.test.id
  bot_version                      = "DRAFT"
  n_lu_intent_confidence_threshold = 0.7
}

resource "aws_lexv2models_bot_version" "test" {
  bot_id = aws_lexv2models_bot.test.id
  locale_specification = {
    (aws_lexv2models_bot_locale.test.locale_id) = {
      source_bot_version = "DRAFT"
    }
  }
}
resource "aws_lexv2models_slot_type" "test" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = "test"
  locale_id   = aws_lexv2models_bot_locale.test.locale_id
}
```

## Argument Reference

The following arguments are required:

* `bot_id` - (Required) Identifier of the bot associated with this slot type.
* `bot_version` - (Required) Version of the bot associated with this slot type.
* `locale_id` - (Required) Identifier of the language and locale where this slot type is used. All of the bots, slot types, and slots used by the intent must have the same locale.
* `name` - (Required) Name of the slot type

The following arguments are optional:

* `description` - (Optional) Description of the slot type.
* `composite_slot_type_setting` - (Optional) Specifications for a composite slot type. See [`composite_slot_type_setting` argument reference](#composite_slot_type_setting-argument-reference) below.
* `external_source_setting` - (Optional) Type of external information used to create the slot type. See [`external_source_setting` argument reference](#external_source_setting-argument-reference) below.
* `parent_slot_type_signature` - (Optional) Built-in slot type used as a parent of this slot type. When you define a parent slot type, the new slot type has the configuration of the parent slot type. Only AMAZON.AlphaNumeric is supported.
* `slot_type_values` - (Optional) List of SlotTypeValue objects that defines the values that the slot type can take. Each value can have a list of synonyms, additional values that help train the machine learning model about the values that it resolves for a slot. See [`slot_type_values` argument reference](#slot_type_values-argument-reference) below.
* `value_selection_setting` - (Optional) Determines the strategy that Amazon Lex uses to select a value from the list of possible values. The field can be set to one of the following values: `ORIGINAL_VALUE` returns the value entered by the user, if the user value is similar to the slot value. `TOP_RESOLUTION` if there is a resolution list for the slot, return the first value in the resolution list. If there is no resolution list, return null. If you don't specify the valueSelectionSetting parameter, the default is ORIGINAL_VALUE. See [`value_selection_setting` argument reference](#value_selection_setting-argument-reference) below.

### `slot_type_values` Argument Reference

* `sample_value` - (Optional) Value of the slot type entry.  See [`sample_value` argument reference](#sample_value-argument-reference) below.
* `synonyms` - (Optional) Additional values related to the slot type entry. See [`sample_value` argument reference](#sample_value-argument-reference) below.

### `sample_value` Argument Reference

* `value` - (Required) Value that can be used for a slot type.

### `external_source_setting` Argument Reference

*`grammar_slot_type_setting` - (Optional) Settings required for a slot type based on a grammar that you provide. See [`grammar_slot_type_setting` argument reference](#grammar_slot_type_setting-argument-reference) below.

### `grammar_slot_type_setting` Argument Reference

* `source` - (Optional) Source of the grammar used to create the slot type. See [`grammar_slot_type_source` argument reference](#grammar_slot_type_source-argument-reference) below.

### `grammar_slot_type_source` Argument Reference

* `s3_bucket_name` - (Required) Name of the Amazon S3 bucket that contains the grammar source.
* `s3_object_key` - (Required) Path to the grammar in the Amazon S3 bucket.
* `kms_key_arn` - (Optional) KMS key required to decrypt the contents of the grammar, if any.

### `composite_slot_type_setting` Argument Reference

* `sub_slots` - (Optional) Subslots in the composite slot. Contains filtered or unexported fields. See [`sub_slot_type_composition` argument reference] below.

### `sub_slot_type_composition` Argument Reference

* `name` - (Required) Name of a constituent sub slot inside a composite slot.
* `slot_type_id` - (Required) Unique identifier assigned to a slot type. This refers to either a built-in slot type or the unique slotTypeId of a custom slot type.

### `value_selection_setting` Argument Reference

* `resolution_strategy` - (Required) Determines the slot resolution strategy that Amazon Lex uses to return slot type values. The field can be set to one of the following values: `ORIGINAL_VALUE` - Returns the value entered by the user, if the user value is similar to the slot value. `TOP_RESOLUTION` If there is a resolution list for the slot, return the first value in the resolution list as the slot type value. If there is no resolution list, null is returned. If you don't specify the valueSelectionStrategy , the default is `ORIGINAL_VALUE`. Valid values are `OriginalValue`, `TopResolution`, and `Concatenation`.
* `advanced_recognition_setting` - (Optional) Provides settings that enable advanced recognition settings for slot values. You can use this to enable using slot values as a custom vocabulary for recognizing user utterances. See [`advanced_recognition_setting` argument reference] below.
* `regex_filter` - (Optional) Used to validate the value of the slot. See [`regex_filter` argument reference] below.

### `advanced_recognition_setting` Argument Reference

* `pattern` - (Required) Used to validate the value of a slot. Use a standard regular expression. Amazon Lex supports the following characters in the regular expression: A-Z, a-z, 0-9, Unicode characters ("\⁠u").
Represent Unicode characters with four digits, for example "\⁠u0041" or "\⁠u005A". The following regular expression operators are not supported: Infinite repeaters: *, +, or {x,} with no upper bound, wild card (.)

### `advanced_recognition_setting` Argument Reference

* `audio_recognition_strategy` - (Optional) Enables using the slot values as a custom vocabulary for recognizing user utterances. Valid value is `UseSlotValuesAsCustomVocabulary`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `id` - Comma-delimited string concatenating `bot_id`, `bot_version`, `locale_id`, and `slot_type_id`.
* `slot_id` - Unique identifier for the intent.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `60m`)
* `update` - (Default `180m`)
* `delete` - (Default `90m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Slot Type using the `example_id_arg`. For example:

```terraform
import {
  to = aws_lexv2models_slot_type.example
  id = "slot_type-id-12345678"
}
```

Using `terraform import`, import Lex V2 Models Slot Type using the `example_id_arg`. For example:

```console
% terraform import aws_lexv2models_slot_type.example bot-1234,DRAFT,en_US,slot_type-id-12345678
```
