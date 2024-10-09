---
subcategory: "Lex V2 Models"
layout: "aws"
page_title: "AWS: aws_lexv2models_intent"
description: |-
  Terraform resource for managing an AWS Lex V2 Models Intent.
---

# Resource: aws_lexv2models_intent

Terraform resource for managing an AWS Lex V2 Models Intent.

## Example Usage

### Basic Usage

```terraform
data "aws_partition" "current" {}

resource "aws_iam_role" "test" {
  name = "botens_namn"
  assume_role_policy = jsonencode({
    Version = "2012-10-17"
    Statement = [
      {
        Action = "sts:AssumeRole"
        Effect = "Allow"
        Sid    = ""
        Principal = {
          Service = "lexv2.amazonaws.com"
        }
      },
    ]
  })
}

resource "aws_iam_role_policy_attachment" "test" {
  role       = aws_iam_role.test.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonLexFullAccess"
}

resource "aws_lexv2models_bot" "test" {
  name                        = "botens_namn"
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

resource "aws_lexv2models_intent" "example" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = "botens_namn"
  locale_id   = aws_lexv2models_bot_locale.test.locale_id
}
```

### `confirmation_setting` Example

When using `confirmation_setting`, if you do not provide a `prompt_attempts_specification`, AWS Lex will provide default `prompt_attempts_specification`s. As a result, Terraform will report a difference in the configuration. To avoid this behavior, include the default `prompt_attempts_specification` configuration shown below.

```terraform
resource "aws_lexv2models_intent" "example" {
  bot_id      = aws_lexv2models_bot.test.id
  bot_version = aws_lexv2models_bot_locale.test.bot_version
  name        = "botens_namn"
  locale_id   = aws_lexv2models_bot_locale.test.locale_id

  confirmation_setting {
    active = true

    prompt_specification {
      allow_interrupt            = true
      max_retries                = 1
      message_selection_strategy = "Ordered"

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

* `bot_id` - (Required) Identifier of the bot associated with this intent.
* `bot_version` - (Required) Version of the bot associated with this intent.
* `locale_id` - (Required) Identifier of the language and locale where this intent is used. All of the bots, slot types, and slots used by the intent must have the same locale.
* `name` - (Required) Name of the intent. Intent names must be unique in the locale that contains the intent and cannot match the name of any built-in intent.

The following arguments are optional:

* `closing_setting` - (Optional) Configuration block for the response that Amazon Lex sends to the user when the intent is closed. See [`closing_setting`](#closing_setting).
* `confirmation_setting` - (Optional) Configuration block for prompts that Amazon Lex sends to the user to confirm the completion of an intent. If the user answers "no," the settings contain a statement that is sent to the user to end the intent. If you configure this block without `prompt_specification.*.prompt_attempts_specification`, AWS will provide default configurations for `Initial` and `Retry1` `prompt_attempts_specification`s. This will cause Terraform to report differences. Use the `confirmation_setting` configuration above in the [Basic Usage](#basic-usage) example to avoid differences resulting from AWS default configuration. See [`confirmation_setting`](#confirmation_setting).
* `description` - (Optional) Description of the intent. Use the description to help identify the intent in lists.
* `dialog_code_hook` - (Optional) Configuration block for invoking the alias Lambda function for each user input. You can invoke this Lambda function to personalize user interaction. See [`dialog_code_hook`](#dialog_code_hook).
* `fulfillment_code_hook` - (Optional) Configuration block for invoking the alias Lambda function when the intent is ready for fulfillment. You can invoke this function to complete the bot's transaction with the user. See [`fulfillment_code_hook`](#fulfillment_code_hook).
* `initial_response_setting` - (Optional) Configuration block for the response that is sent to the user at the beginning of a conversation, before eliciting slot values. See [`initial_response_setting`](#initial_response_setting).
* `input_context` - (Optional) Configuration blocks for contexts that must be active for this intent to be considered by Amazon Lex. When an intent has an input context list, Amazon Lex only considers using the intent in an interaction with the user when the specified contexts are included in the active context list for the session. If the contexts are not active, then Amazon Lex will not use the intent. A context can be automatically activated using the outputContexts property or it can be set at runtime. See [`input_context`](#input_context).
* `kendra_configuration` - (Optional) Configuration block for information required to use the AMAZON.KendraSearchIntent intent to connect to an Amazon Kendra index. The AMAZON.KendraSearchIntent intent is called when Amazon Lex can't determine another intent to invoke. See [`kendra_configuration`](#kendra_configuration).
* `output_context` - (Optional) Configuration blocks for contexts that the intent activates when it is fulfilled. You can use an output context to indicate the intents that Amazon Lex should consider for the next turn of the conversation with a customer. When you use the outputContextsList property, all of the contexts specified in the list are activated when the intent is fulfilled. You can set up to 10 output contexts. You can also set the number of conversation turns that the context should be active, or the length of time that the context should be active. See [`output_context`](#output_context).
* `parent_intent_signature` - (Optional) Identifier for the built-in intent to base this intent on.
* `sample_utterance` - (Optional) Configuration block for strings that a user might say to signal the intent. See [`sample_utterance`](#sample_utterance).
* `slot_priority` - (Optional) Configuration block for a new list of slots and their priorities that are contained by the intent. This is ignored on create and only valid for updates. See [`slot_priority`](#slot_priority).

### `closing_setting`

* `active` - (Optional) Whether an intent's closing response is used. When this field is false, the closing response isn't sent to the user. If the active field isn't specified, the default is true.
* `closing_response` - (Optional) Configuration block for response that Amazon Lex sends to the user when the intent is complete. See [`closing_response`](#closing_response).
* `conditional` - (Optional) Configuration block for list of conditional branches associated with the intent's closing response. These branches are executed when the `next_step` attribute is set to `EvalutateConditional`. See [`conditional`](#conditional).
* `next_step` - (Optional) Next step that the bot executes after playing the intent's closing response. See [`next_step`](#next_step).

#### `closing_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

##### `message_group`

* `message` - (Required) Configuration block for the primary message that Amazon Lex should send to the user. See [`message`](#message-and-variation).
* `variation` - (Optional) Configuration blocks for message variations to send to the user. When variations are defined, Amazon Lex chooses the primary message or one of the variations to send to the user. See [`variation`](#message-and-variation).

###### `message` and `variation`

* `custom_payload` - (Optional) Configuration block for a message in a custom format defined by the client application. See [`custom_payload`](#custom_payload).
* `image_response_card` - (Optional) Configuration block for a message that defines a response card that the client application can show to the user. See [`image_response_card`](#image_response_card).
* `plain_text_message` - (Optional) Configuration block for a message in plain text format. See [`plain_text_message`](#plain_text_message).
* `ssml_message` - (Optional) Configuration block for a message in Speech Synthesis Markup Language (SSML). See [`ssml_message`](#ssml_message).

###### `custom_payload`

* `value` - (Required) String that is sent to your application.

###### `image_response_card`

* `title` - (Required) Title to display on the response card. The format of the title is determined by the platform displaying the response card.
* `button` - (Optional) Configuration blocks for buttons that should be displayed on the response card. The arrangement of the buttons is determined by the platform that displays the button. See [`button`](#button).
* `image_url` - (Optional) URL of an image to display on the response card. The image URL must be publicly available so that the platform displaying the response card has access to the image.
* `subtitle` - (Optional) Subtitle to display on the response card. The format of the subtitle is determined by the platform displaying the response card.

###### `button`

* `text` - (Required) Text that appears on the button. Use this to tell the user what value is returned when they choose this button.
* `value` - (Required) Value returned to Amazon Lex when the user chooses this button. This must be one of the slot values configured for the slot.

###### `plain_text_message`

* `value` - (Required) Message to send to the user.

###### `ssml_message`

* `value` - (Required) SSML text that defines the prompt.

#### `conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

##### `conditional_branch`

* `condition` - (Required) Configuration block for the expression to evaluate. If the condition is true, the branch's actions are taken. See [`condition`](#condition).
* `name` - (Required) Name of the branch.
* `next_step` - (Required) Configuration block for the next step in the conversation. See [`next_step`](#next_step).
* `response` - (Optional) Configuration block for a list of message groups that Amazon Lex uses to respond to the user input. See [`response`](#response).

###### `condition`

* `expression_string` - (Required) Expression string that is evaluated.

###### `response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

###### `default_branch`

* `next_step` - (Required) Configuration block for the next step in the conversation. See [`next_step`](#next_step).
* `response` - (Optional) Configuration block for a list of message groups that Amazon Lex uses to respond to the user input. See [`response`](#response).

#### `next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

##### `dialog_action`

* `type` - (Required) Action that the bot should execute. Valid values are `ElicitIntent`, `StartIntent`, `ElicitSlot`, `EvaluateConditional`, `InvokeDialogCodeHook`, `ConfirmIntent`, `FulfillIntent`, `CloseIntent`, `EndConversation`.
* `slot_to_elicit` - (Optional) If the dialog action is `ElicitSlot`, defines the slot to elicit from the user.
* `suppress_next_message` - (Optional) Whether the next message for the intent is _not_ used.

##### `intent`

* `name` - (Optional, Required when switching intents) Name of the intent.
* `slot` - (Optional) Configuration block for all of the slot value overrides for the intent. The name of the slot maps to the value of the slot. Slots that are not included in the map aren't overridden. See [`slot`](#slot).

###### `slot`

* `shape` - (Optional) When the shape value is `List`, `values` contains a list of slot values. When the value is `Scalar`, `value` contains a single value.
* `value` - (Optional) Configuration block for the current value of the slot. See [`value`](#slot-value).
* `values` - _Not currently supported._

###### Slot `value`

* `interpreted_value` - (Optional) Value that Amazon Lex determines for the slot. The actual value depends on the setting of the value selection strategy for the bot. You can choose to use the value entered by the user, or you can have Amazon Lex choose the first value in the resolvedValues list.

### `confirmation_setting`

* `prompt_specification` - (Required) Configuration block for prompting the user to confirm the intent. This question should have a yes or no answer. Amazon Lex uses this prompt to ensure that the user acknowledges that the intent is ready for fulfillment. See [`prompt_specification`](#prompt_specification).
* `active` - (Optional) Whether the intent's confirmation is sent to the user. When this field is false, confirmation and declination responses aren't sent. If the active field isn't specified, the default is true.
* `code_hook` - (Optional) Configuration block for the intent's confirmation step. The dialog code hook is triggered based on these invocation settings when the confirmation next step or declination next step or failure next step is `invoke_dialog_code_hook`.  See [`code_hook`](#code_hook).
* `confirmation_conditional` - (Optional) Configuration block for conditional branches to evaluate after the intent is closed. See [`confirmation_conditional`](#confirmation_conditional).
* `confirmation_next_step` - (Optional) Configuration block for the next step that the bot executes when the customer confirms the intent. See [`confirmation_next_step`](#confirmation_next_step).
* `confirmation_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`confirmation_response`](#confirmation_response).
* `declination_conditional` - (Optional) Configuration block for conditional branches to evaluate after the intent is declined. See [`declination_conditional`](#declination_conditional).
* `declination_next_step` - (Optional) Configuration block for the next step that the bot executes when the customer declines the intent. See [`declination_next_step`](#declination_next_step).
* `declination_response` - (Optional) Configuration block for when the user answers "no" to the question defined in `prompt_specification`, Amazon Lex responds with this response to acknowledge that the intent was canceled. See [`declination_response`](#declination_response).
* `elicitation_code_hook` - (Optional) Configuration block for when the code hook is invoked during confirmation prompt retries. See [`elicitation_code_hook`](#elicitation_code_hook).
* `failure_conditional` - (Optional) Configuration block for conditional branches. Branches are evaluated in the order that they are entered in the list. The first branch with a condition that evaluates to true is executed. The last branch in the list is the default branch. The default branch should not have any condition expression. The default branch is executed if no other branch has a matching condition. See [`failure_conditional`](#failure_conditional).
* `failure_next_step` - (Optional) Configuration block for the next step to take in the conversation if the confirmation step fails. See [`failure_next_step`](#failure_next_step).
* `failure_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`failure_response`](#failure_response).

#### `prompt_specification`

* `max_retries` - (Required) Maximum number of times the bot tries to elicit a response from the user using this prompt.
* `message_group` - (Required) Configuration block for messages that Amazon Lex can send to the user. Amazon Lex chooses the actual message to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech prompt from the bot.
* `message_selection_strategy` - (Optional) How a message is selected from a message group among retries. Valid values are `Random` and `Ordered`.
* `prompt_attempts_specification` - (Optional) Configuration block for advanced settings on each attempt of the prompt. See [`prompt_attempts_specification`](#prompt_attempts_specification).

##### `prompt_attempts_specification`

* `allowed_input_types` - (Required) Configuration block for the allowed input types of the prompt attempt. See [`allowed_input_types`](#allowed_input_types).
* `map_block_key` - (Required) Which attempt to configure. Valid values are `Initial`, `Retry1`, `Retry2`, `Retry3`, `Retry4`, `Retry5`.
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech prompt attempt from the bot.
* `audio_and_dtmf_input_specification` - (Optional) Configuration block for settings on audio and DTMF input. See [`audio_and_dtmf_input_specification`](#audio_and_dtmf_input_specification).
* `text_input_specification` - (Optional) Configuration block for the settings on text input. See [`text_input_specification`](#text_input_specification).

###### `allowed_input_types`

* `allow_audio_input` - (Required) Whether audio input is allowed.
* `allow_dtmf_input` - (Required) Whether DTMF input is allowed.

###### `audio_and_dtmf_input_specification`

* `start_timeout_ms` - (Required) Time for which a bot waits before assuming that the customer isn't going to speak or press a key. This timeout is shared between Audio and DTMF inputs.
* `audio_specification` - (Optional) Configuration block for the settings on audio input. See [`audio_specification`](#audio_specification).
* `dtmf_specification` - (Optional) Configuration block for the settings on DTMF input. See [`dtmf_specification`](#dtmf_specification).

###### `audio_specification`

* `end_timeout_ms` - (Required) Time for which a bot waits after the customer stops speaking to assume the utterance is finished.
* `max_length_ms` - (Required) Time for how long Amazon Lex waits before speech input is truncated and the speech is returned to application.

###### `dtmf_specification`

* `deletion_character` - (Required) DTMF character that clears the accumulated DTMF digits and immediately ends the input.
* `end_character` - (Required) DTMF character that immediately ends input. If the user does not press this character, the input ends after the end timeout.
* `end_timeout_ms` - (Required) How long the bot should wait after the last DTMF character input before assuming that the input has concluded.
* `max_length` - (Required) Maximum number of DTMF digits allowed in an utterance.

###### `text_input_specification`

* `start_timeout_ms` - (Required) Time for which a bot waits before re-prompting a customer for text input.

### `code_hook`

* `active` - (Required) Whether a dialog code hook is used when the intent is activated.
* `enable_code_hook_invocation` - (Required) Whether a Lambda function should be invoked for the dialog.
* `post_code_hook_specification` - (Required) Configuration block that contains the responses and actions that Amazon Lex takes after the Lambda function is complete. See [`post_code_hook_specification`](#post_code_hook_specification).
* `invocation_label` - (Optional) Label that indicates the dialog step from which the dialog code hook is happening.

#### `post_code_hook_specification`

* `failure_conditional` - (Optional) Configuration block for conditional branches to evaluate after the dialog code hook throws an exception or returns with the State field of the Intent object set to Failed.
* `failure_next_step` - (Optional) Configuration block for the next step the bot runs after the dialog code hook throws an exception or returns with the State field of the Intent object set to Failed . See [`failure_next_step`](#failure_next_step).
* `failure_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`failure_response`](#failure_response).
* `success_conditional` - (Optional) Configuration block for conditional branches to evaluate after the dialog code hook finishes successfully. See [`success_conditional`](#success_conditional).
* `success_next_step` - (Optional) Configuration block for the next step the bot runs after the dialog code hook finishes successfully. See [`success_next_step`](#success_next_step).
* `success_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`success_response`](#success_response).
* `timeout_conditional` - (Optional) Configuration block for conditional branches to evaluate if the code hook times out. See [`timeout_conditional`](#timeout_conditional).
* `timeout_next_step` - (Optional) Configuration block for the next step that the bot runs when the code hook times out. See [`timeout_next_step`](#timeout_next_step).
* `timeout_response` - (Optional) Configuration block for a list of message groups that Amazon Lex uses to respond the user input. See [`timeout_response`](#timeout_response).

##### `failure_conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

##### `failure_next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

##### `failure_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

##### `success_conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

##### `success_next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

##### `success_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

##### `timeout_conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

##### `timeout_next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

##### `timeout_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

#### `confirmation_conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

#### `confirmation_next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

#### `confirmation_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

#### `declination_conditional`

* `active` - (Required) Whether a conditional branch is active. When active is false, the conditions are not evaluated.
* `conditional_branch` - (Required) Configuration blocks for conditional branches. A conditional branch is made up of a condition, a response and a next step. The response and next step are executed when the condition is true. See [`conditional_branch`](#conditional_branch).
* `default_branch` - (Required) Configuration block for the conditional branch that should be followed when the conditions for other branches are not satisfied. A branch is made up of a condition, a response and a next step. See [`default_branch`](#default_branch).

#### `declination_next_step`

* `dialog_action` - (Optional) Configuration block for action that the bot executes at runtime when the conversation reaches this step. See [`dialog_action`](#dialog_action).
* `intent` - (Optional) Configuration block for override settings to configure the intent state. See [`intent`](#intent).
* `session_attributes` - (Optional) Map of key/value pairs representing session-specific context information. It contains application information passed between Amazon Lex and a client application.

#### `declination_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

#### `elicitation_code_hook`

* `enable_code_hook_invocation` - (Required) Whether a Lambda function should be invoked for the dialog.
* `invocation_label` - (Optional) Label that indicates the dialog step from which the dialog code hook is happening.

### `dialog_code_hook`

* `enabled` - (Required) Enables the dialog code hook so that it processes user requests.

### `fulfillment_code_hook`

* `enabled` - (Required) Whether a Lambda function should be invoked to fulfill a specific intent.
* `active` - (Optional) Whether the fulfillment code hook is used. When active is false, the code hook doesn't run.
* `fulfillment_updates_specification` - (Optional) Configuration block for settings for update messages sent to the user for long-running Lambda fulfillment functions. Fulfillment updates can be used only with streaming conversations. See [`fulfillment_updates_specification`](#fulfillment_updates_specification).
* `post_fulfillment_status_specification` - (Optional) Configuration block for settings for messages sent to the user for after the Lambda fulfillment function completes. Post-fulfillment messages can be sent for both streaming and non-streaming conversations. See [`post_fulfillment_status_specification`](#post_fulfillment_status_specification).

#### `fulfillment_updates_specification`

* `active` - (Required) Whether fulfillment updates are sent to the user. When this field is true, updates are sent. If the active field is set to true, the `start_response`, `update_response`, and `timeout_in_seconds` fields are required.
* `start_response` - (Required, if `active`) Configuration block for the message sent to users when the fulfillment Lambda functions starts running.
* `timeout_in_seconds` - (Required, if `active`) Length of time that the fulfillment Lambda function should run before it times out.
* `update_response` - (Required, if `active`) Configuration block for messages sent periodically to the user while the fulfillment Lambda function is running.

##### `start_response`

* `delay_in_seconds` - (Required) Delay between when the Lambda fulfillment function starts running and the start message is played. If the Lambda function returns before the delay is over, the start message isn't played.
* `message_group` - (Required) Between 1-5 configuration block message groups that contain start messages. Amazon Lex chooses one of the messages to play to the user. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt the start message while it is playing.

##### `update_response`

* `frequency_in_seconds` - (Required) Frequency that a message is sent to the user. When the period ends, Amazon Lex chooses a message from the message groups and plays it to the user. If the fulfillment Lambda returns before the first period ends, an update message is not played to the user.
* `message_group` - (Required) Between 1-5 configuration block message groups that contain start messages. Amazon Lex chooses one of the messages to play to the user. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt the start message while it is playing.

#### `post_fulfillment_status_specification`

* `failure_conditional` - (Optional) Configuration block for conditional branches to evaluate after the dialog code hook throws an exception or returns with the State field of the Intent object set to Failed. See [`failure_conditional`](#failure_conditional).
* `failure_next_step` - (Optional) Configuration block for the next step the bot runs after the dialog code hook throws an exception or returns with the State field of the Intent object set to Failed. See [`failure_next_step`](#failure_next_step).
* `failure_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`failure_response`](#failure_response).
* `success_conditional` - (Optional) Configuration block for conditional branches to evaluate after the dialog code hook finishes successfully. See [`success_conditional`](#success_conditional).
* `success_next_step` - (Optional) Configuration block for the next step the bot runs after the dialog code hook finishes successfully. See [`success_next_step`](#success_next_step).
* `success_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`success_response`](#success_response).
* `timeout_conditional` - (Optional) Configuration block for conditional branches to evaluate if the code hook times out. See [`timeout_conditional`](#timeout_conditional).
* `timeout_next_step` - (Optional) Configuration block for the next step that the bot runs when the code hook times out. See [`timeout_next_step`](#timeout_next_step).
* `timeout_response` - (Optional) Configuration block for a list of message groups that Amazon Lex uses to respond the user input. See [`timeout_response`](#timeout_response).

### `initial_response_setting`

* `code_hook` - (Optional) Configuration block for the dialog code hook that is called by Amazon Lex at a step of the conversation. See [`code_hook`](#code_hook).
* `conditional` - (Optional) Configuration block for conditional branches. Branches are evaluated in the order that they are entered in the list. The first branch with a condition that evaluates to true is executed. The last branch in the list is the default branch. The default branch should not have any condition expression. The default branch is executed if no other branch has a matching condition. See [`conditional`](#conditional).
* `initial_response` - (Optional) Configuration block for message groups that Amazon Lex uses to respond the user input. See [`initial_response`](#initial_response).
* `next_step` - (Optional) Configuration block for the next step in the conversation. See [`next_step`](#next_step).

#### `initial_response`

* `message_group` - (Required) Configuration blocks for responses that Amazon Lex can send to the user. Amazon Lex chooses the actual response to send at runtime. See [`message_group`](#message_group).
* `allow_interrupt` - (Optional) Whether the user can interrupt a speech response from Amazon Lex.

### `input_context`

* `name` - (Required) Name of the context.

### `kendra_configuration`

* `kendra_index` - (Required) ARN of the Amazon Kendra index that you want the AMAZON.KendraSearchIntent intent to search. The index must be in the same account and Region as the Amazon Lex bot.
* `query_filter_string` - (Optional) Query filter that Amazon Lex sends to Amazon Kendra to filter the response from a query. The filter is in the format defined by Amazon Kendra. For more information, see [Filtering queries](https://docs.aws.amazon.com/kendra/latest/dg/filtering.html).
* `query_filter_string_enabled` - (Optional) Whether the AMAZON.KendraSearchIntent intent uses a custom query string to query the Amazon Kendra index.

### `output_context`

* `name` - (Required) Name of the output context.
* `time_to_live_in_seconds` - (Required) Amount of time, in seconds, that the output context should remain active. The time is figured from the first time the context is sent to the user.
* `turns_to_live` - (Required) Number of conversation turns that the output context should remain active. The number of turns is counted from the first time that the context is sent to the user.

### `sample_utterance`

* `utterance` - (Required) Sample utterance that Amazon Lex uses to build its machine-learning model to recognize intents.

### `slot_priority`

* `priority` - (Required) Priority that Amazon Lex should apply to the slot.
* `slot_id` - (Required) Unique identifier of the slot.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `creation_date_time` - Timestamp of the date and time that the intent was created.
* `id` - Composite identifier of `intent_id:bot_id:bot_version:locale_id`.
* `intent_id` - Unique identifier for the intent.
* `last_updated_date_time` - Timestamp of the last time that the intent was modified.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Lex V2 Models Intent using the `intent_id:bot_id:bot_version:locale_id`. For example:

```terraform
import {
  to = aws_lexv2models_intent.example
  id = "intent-42874:bot-11376:DRAFT:en_US"
}
```

Using `terraform import`, import Lex V2 Models Intent using the `intent_id:bot_id:bot_version:locale_id`. For example:

```console
% terraform import aws_lexv2models_intent.example intent-42874:bot-11376:DRAFT:en_US
```
