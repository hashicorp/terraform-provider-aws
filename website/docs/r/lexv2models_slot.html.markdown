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

### `value_elicitation_setting` Example

~> When using `value_elicitation_setting`, if you do not provide a `prompt_attempts_specification`, AWS Lex will configure default `prompt_attempts_specification`s.
As a result, Terraform will report a difference in the configuration.
To avoid this behavior, include `prompt_attempts_specification` blocks matching the default configuration, as shown below.

```terraform
resource "aws_lexv2models_slot" "example" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  intent_id   = aws_lexv2models_intent.test.intent_id
  locale_id   = aws_lexv2models_bot_locale.test.locale_id
  name        = "example"

  value_elicitation_setting {
    slot_constraint = "Required"
    prompt_specification {
      allow_interrupt            = true
      max_retries                = 1
      message_selection_strategy = "Random"

      message_group {
        message {
          plain_text_message {
            value = "What is your favorite color?"
          }
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Initial"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = 640
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }

      prompt_attempts_specification {
        allow_interrupt = true
        map_block_key   = "Retry1"

        allowed_input_types {
          allow_audio_input = true
          allow_dtmf_input  = true
        }

        audio_and_dtmf_input_specification {
          start_timeout_ms = 4000

          audio_specification {
            end_timeout_ms = 640
            max_length_ms  = 15000
          }

          dtmf_specification {
            deletion_character = "*"
            end_character      = "#"
            end_timeout_ms     = 5000
            max_length         = 513
          }
        }

        text_input_specification {
          start_timeout_ms = 30000
        }
      }

    }
  }
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
If you configure this block without `prompt_specification.*.prompt_attempts_specification`, AWS will provide default `prompt_attempts_specification` blocks for the initial prompt (map key `Initial`) and each retry attempt (map keys `Retry1`, `Retry2`, etc.).
This will cause Terraform to report differences.
Use the `value_elicitation_setting` configuration above in the [`value_elicitation_setting` example](#value_elicitation_setting-example) to avoid differences resulting from AWS default configurations.
See the [`value_elicitation_setting` argument reference](#value_elicitation_setting-argument-reference) below.

The following arguments are optional:

* `description` - (Optional) Description of the slot.
* `multiple_values_setting` - (Optional) Whether the slot returns multiple values in one response.
See the [`multiple_values_setting` argument reference](#multiple_values_setting-argument-reference) below.
* `obfuscation_setting` - (Optional) Determines how slot values are used in Amazon CloudWatch logs.
See the [`obfuscation_setting` argument reference](#obfuscation_setting-argument-reference) below.
* `slot_type_id` - (Optional) Unique identifier for the slot type associated with this slot.
* `sub_slot_setting` - (Optional) Specifications for the constituent sub slots and the expression for the composite slot.
See the [`sub_slot_setting` argument reference](#sub_slot_setting-argument-reference) below.

### `multiple_values_setting` Argument Reference

* `allow_multiple_values` - (Optional) Whether a slot can return multiple values. When `true`, the slot may return more than one value in a response. When `false`, the slot returns only a single value. Multi-value slots are only available in the `en-US` locale.

### `obfuscation_setting` Argument Reference

* `obfuscation_setting_type` - (Required) Whether Amazon Lex obscures slot values in conversation logs. Valid values are `DefaultObfuscation` and `None`.

### `sub_slot_setting` Argument Reference

* `expression` - (Optional) Expression text for defining the constituent sub slots in the composite slot using logical `AND` and `OR` operators.
* `slot_specification` - (Optional) Specifications for the constituent sub slots of a composite slot.
See the [`slot_specification` argument reference](#slot_specification-argument-reference) below.

#### `slot_specification` Argument Reference

* `slot_type_id` - (Required) Unique identifier assigned to the slot type.
* `value_elicitation_setting` - (Required) Elicitation setting details for constituent sub slots of a composite slot.
See the [`value_elicitation_setting` argument reference](#value_elicitation_setting-argument-reference) below.

### `value_elicitation_setting` Argument Reference

* `slot_constraint` - (Required) Whether the slot is required or optional. Valid values are `Required` or `Optional`.
* `default_value_specification` - (Optional) List of default values for a slot.
See the [`default_value_specification` argument reference](#default_value_specification-argument-reference) below.
* `prompt_specification` - (Optional) Prompt that Amazon Lex uses to elicit the slot value from the user.
See the [`aws_lexv2models_intent` resource](/docs/providers/aws/r/lexv2models_intent.html) for details on the `prompt_specification` argument reference - they are identical.
* `sample_utterances` - (Optional) A specific pattern that users might respond to an Amazon Lex request for a slot value.
See the [`sample_utterances` argument reference](#sample_utterances-argument-reference) below.
* `slot_resolution_setting` - (Optional) Information about whether assisted slot resolution is turned on for the slot or not.
See the [`slot_resolution_setting` argument reference](#slot_resolution_setting-argument-reference) below.
* `wait_and_continue_specification` - (Optional) Specifies the prompts that Amazon Lex uses while a bot is waiting for customer input.
See the [`wait_and_continue_specification` argument reference](#wait_and_continue_specification-argument-reference) below.

#### `default_value_specification` Argument Reference

* `default_value_list` - (Required) List of default values.
Amazon Lex chooses the default value to use in the order that they are presented in the list.
See the [`default_value_list` argument reference](#default_value_list-argument-reference) below.

##### `default_value_list` Argument Reference

* `default_value` - (Required) Default value to use when a user doesn't provide a value for a slot.

#### `sample_utterances` Argument Reference

* `utterance` - (Required) The sample utterance that Amazon Lex uses to build its machine-learning model to recognize intents.

#### `slot_resolution_setting` Argument Reference

* `slot_resolution_strategy` - (Required) Specifies whether assisted slot resolution is turned on for the slot or not.
Valid values are `EnhancedFallback` or `Default`.
If the value is `EnhancedFallback`, assisted slot resolution is activated when Amazon Lex defaults to the `AMAZON.FallbackIntent`.
If the value is `Default`, assisted slot resolution is turned off.

#### `wait_and_continue_specification` Argument Reference

* `continue_response` - (Required) Response that Amazon Lex sends to indicate that the bot is ready to continue the conversation.
See the [`continue_response` argument reference](#continue_response-argument-reference) below.
* `waiting_response` - (Required) Response that Amazon Lex sends to indicate that the bot is waiting for the conversation to continue.
See the [`waiting_response` argument reference](#waiting_response-argument-reference) below.
* `active` - (Optional) Specifies whether the bot will wait for a user to respond.
When this field is `false`, wait and continue responses for a slot aren't used.
If the active field isn't specified, the default is `true`.
* `still_waiting_response` - (Optional) Response that Amazon Lex sends periodically to the user to indicate that the bot is still waiting for input from the user.
See the [`still_waiting_response` argument reference](#still_waiting_response-argument-reference) below.

##### `continue_response` Argument Reference

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user.
Amazon Lex chooses the actual response to send at runtime.
See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

##### `waiting_response` Argument Reference

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user.
Amazon Lex chooses the actual response to send at runtime.
See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

##### `still_waiting_response` Argument Reference

* `frequency_in_seconds` - (Required) How often a message should be sent to the user.
* `message_groups` - (Required) One or more message groups, each containing one or more messages, that define the prompts that Amazon Lex sends to the user.
See [`message_group`](#message_group).
* `timeout_in_seconds` - (Required) If Amazon Lex waits longer than this length of time for a response, it will stop sending messages.
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

###### `message_group`

* `message` - (Required) Configuration block for the primary message that Amazon Lex should send to the user.
See the [`aws_lexv2models_intent` resource](/docs/providers/aws/r/lexv2models_intent.html) for details on the `message` argument reference - they are identical.
* `variation` - (Optional) Configuration blocks for message variations to send to the user.
When variations are defined, Amazon Lex chooses the primary message or one of the variations to send to the user.
See the [`aws_lexv2models_intent` resource](/docs/providers/aws/r/lexv2models_intent.html) for details on the `variation` argument reference - they are identical.

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
