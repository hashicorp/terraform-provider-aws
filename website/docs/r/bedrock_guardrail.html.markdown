---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail"
description: |-
  Terraform resource for managing an Amazon Bedrock Guardrail.
---

# Resource: aws_bedrock_guardrail

Terraform resource for managing an Amazon Bedrock Guardrail.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "example"
  blocked_input_messaging   = "example"
  blocked_outputs_messaging = "example"
  description               = "example"

  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
    tier_config {
      tier_name = "STANDARD"
    }
  }

  sensitive_information_policy_config {
    pii_entities_config {
      action         = "BLOCK"
      input_action   = "BLOCK"
      output_action  = "ANONYMIZE"
      input_enabled  = true
      output_enabled = true
      type           = "NAME"
    }

    regexes_config {
      action         = "BLOCK"
      input_action   = "BLOCK"
      output_action  = "BLOCK"
      input_enabled  = true
      output_enabled = false
      description    = "example regex"
      name           = "regex_example"
      pattern        = "^\\d{3}-\\d{2}-\\d{4}$"
    }
  }

  topic_policy_config {
    topics_config {
      name       = "investment_topic"
      examples   = ["Where should I invest my money ?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns ."
    }
    tier_config {
      tier_name = "CLASSIC"
    }
  }

  word_policy_config {
    managed_word_lists_config {
      type = "PROFANITY"
    }
    words_config {
      text = "HATE"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `blocked_input_messaging` - (Required) Message to return when the guardrail blocks a prompt.
* `blocked_outputs_messaging` - (Required) Message to return when the guardrail blocks a model response.
* `name` - (Required) Name of the guardrail.

The following arguments are optional:

* `content_policy_config` - (Optional) Content policy config for a guardrail. See [`content_policy_config` Block](#content_policy_config-block) for more information.
* `contextual_grounding_policy_config` - (Optional) Contextual grounding policy config for a guardrail. See [`contextual_grounding_policy_config` Block](#contextual_grounding_policy_config-block) for more information.
* `cross_region_config` - (Optional) Configuration block to enable cross-region routing for bedrock guardrails. See [`cross_region_config` Block](#cross_region_config-block) for more information. Note see [available regions](https://docs.aws.amazon.com/bedrock/latest/userguide/guardrails-cross-region.html) here.
* `description` - (Optional) Description of the guardrail or its version.
* `kms_key_arn` - (Optional) KMS key with which the guardrail was encrypted at rest.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `sensitive_information_policy_config` - (Optional) Sensitive information policy config for a guardrail. See [`sensitive_information_policy_config` Block](#sensitive_information_policy_config-block) for more information.
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `topic_policy_config` - (Optional) Topic policy config for a guardrail. See [`topic_policy_config` Block](#topic_policy_config-block) for more information.
* `word_policy_config` - (Optional) Word policy config for a guardrail. See [`word_policy_config` Block](#word_policy_config-block) for more information.

### `content_policy_config` Block

The `content_policy_config` configuration block supports the following arguments:

* `filters_config` - (Optional) Set of content filter configs in content policy. See [`content_policy_config.filters_config` Block](#content_policy_configfilters_config-block) for more information.
* `tier_config` - (Optional) Configuration block for the content policy tier. See [`content_policy_config.tier_config` Block](#content_policy_configtier_config-block) for more information.

### `content_policy_config.filters_config` Block

The `filters_config` configuration block supports the following arguments:

* `input_action` - (Optional) Action to take when harmful content is detected. Valid values: `BLOCK`, `NONE`.
* `input_enabled` - (Optional) Toggles guardrail evaluation on input.
* `input_modalities` - (Optional) List of selected input modalities. Valid values: `IMAGE`, `TEXT`.
* `input_strength` - (Required) Strength for filters. Valid values: `NONE`, `LOW`, `MEDIUM`, `HIGH`.
* `output_action` - (Optional) Action to take when harmful content is detected. Valid values: `BLOCK`, `NONE`.
* `output_enabled` - (Optional) Toggles guardrail evaluation on output.
* `output_modalities` - (Optional) List of selected output modalities. Valid values: `IMAGE`, `TEXT`.
* `output_strength` - (Required) Strength for filters. Valid values: `NONE`, `LOW`, `MEDIUM`, `HIGH`.
* `type` - (Required) Type of filter in content policy. Valid values: `SEXUAL`, `VIOLENCE`, `HATE`, `INSULTS`, `MISCONDUCT`, `PROMPT_ATTACK`.

### `content_policy_config.tier_config` Block

The `tier_config` configuration block supports the following arguments:

* `tier_name` - (Required) Name of the content policy tier. Valid values include STANDARD or CLASSIC.

### `contextual_grounding_policy_config` Block

The `contextual_grounding_policy_config` configuration block supports the following arguments:

* `filters_config` - (Required) One or more blocks defining contextual grounding filter configs. See [`contextual_grounding_policy_config.filters_config` Block](#contextual_grounding_policy_configfilters_config-block) for more information.

### `contextual_grounding_policy_config.filters_config` Block

The `filters_config` configuration block supports the following arguments:

* `threshold` - (Required) Threshold for this filter.
* `type` - (Required) Type of contextual grounding filter.

### `cross_region_config` Block

The `cross_region_config` configuration block supports the following arguments:

* `guardrail_profile_identifier` - (Required) Guardrail profile ARN.

### `sensitive_information_policy_config` Block

The `sensitive_information_policy_config` configuration block supports the following arguments:

* `pii_entities_config` - (Optional) List of entities. See [`pii_entities_config` Block](#pii_entities_config-block) for more information.
* `regexes_config` - (Optional) List of regex. See [`regexes_config` Block](#regexes_config-block) for more information.

### `pii_entities_config` Block

The `pii_entities_config` configuration block supports the following arguments:

* `action` - (Required) Options for sensitive information action. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `input_action` - (Optional) Action to take when harmful content is detected in the input. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `input_enabled` - (Optional) Whether to enable guardrail evaluation on the input. When disabled, you aren't charged for the evaluation.
* `output_action` - (Optional) Action to take when harmful content is detected in the output. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `output_enabled` - (Optional) Whether to enable guardrail evaluation on the output. When disabled, you aren't charged for the evaluation.
* `type` - (Required) Currently supported PII entities.

### `regexes_config` Block

The `regexes_config` configuration block supports the following arguments:

* `action` - (Required) Options for sensitive information action. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `description` - (Optional) Regex description.
* `input_action` - (Optional) Action to take when harmful content is detected in the input. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `input_enabled` - (Optional) Whether to enable guardrail evaluation on the input. When disabled, you aren't charged for the evaluation.
* `name` - (Required) Regex name.
* `output_action` - (Optional) Action to take when harmful content is detected in the output. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
* `output_enabled` - (Optional) Whether to enable guardrail evaluation on the output. When disabled, you aren't charged for the evaluation.
* `pattern` - (Required) Regex pattern.

### `topic_policy_config` Block

The `topic_policy_config` configuration block supports the following arguments:

* `tier_config` - (Optional) Configuration block for the topic policy tier. See [`topic_policy_config.tier_config` Block](#topic_policy_configtier_config-block) for more information.
* `topics_config` - (Required) List of topic configs in topic policy. See [`topics_config` Block](#topics_config-block) for more information.

### `topic_policy_config.tier_config` Block

The `tier_config` configuration block supports the following arguments:

* `tier_name` - (Required) Name of the topic policy tier. Valid values include STANDARD or CLASSIC.

### `topics_config` Block

The `topics_config` configuration block supports the following arguments:

* `definition` - (Required) Definition of topic in topic policy.
* `examples` - (Optional) List of text examples.
* `name` - (Required) Name of topic in topic policy.
* `type` - (Required) Type of topic in a policy.

### `word_policy_config` Block

The `word_policy_config` configuration block supports the following arguments:

* `managed_word_lists_config` - (Optional) Config for the list of managed words. See [`managed_word_lists_config` Block](#managed_word_lists_config-block) for more information.
* `words_config` - (Optional) List of custom word configs. See [`words_config` Block](#words_config-block) for more information.

### `managed_word_lists_config` Block

The `managed_word_lists_config` configuration block supports the following arguments:

* `input_action` - (Optional) Action to take when harmful content is detected in the input. Valid values: `BLOCK`, `NONE`.
* `input_enabled` - (Optional) Whether to enable guardrail evaluation on the input. When disabled, you aren't charged for the evaluation.
* `output_action` - (Optional) Action to take when harmful content is detected in the output. Valid values: `BLOCK`, `NONE`.
* `output_enabled` - (Optional) Whether to enable guardrail evaluation on the output. When disabled, you aren't charged for the evaluation.
* `type` - (Required) Options for managed words.

### `words_config` Block

The `words_config` configuration block supports the following arguments:

* `input_action` - (Optional) Action to take when harmful content is detected in the input. Valid values: `BLOCK`, `NONE`.
* `input_enabled` - (Optional) Whether to enable guardrail evaluation on the input. When disabled, you aren't charged for the evaluation.
* `output_action` - (Optional) Action to take when harmful content is detected in the output. Valid values: `BLOCK`, `NONE`.
* `output_enabled` - (Optional) Whether to enable guardrail evaluation on the output. When disabled, you aren't charged for the evaluation.
* `text` - (Required) Custom word text.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `created_at` - Unix epoch timestamp in seconds for when the Guardrail was created.
* `guardrail_arn` - ARN of the Guardrail.
* `guardrail_id` - ID of the Guardrail.
* `status` - Status of the Bedrock Guardrail. One of `READY`, `FAILED`.
* `updated_at` - Date and time that the Guardrail list was last updated.
* `version` - Version of the Guardrail.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Bedrock Guardrail using a comma-delimited string of `guardrail_id` and `version`. For example:

```terraform
import {
  to = aws_bedrock_guardrail.example
  id = "guardrail-id-12345678,DRAFT"
}
```

Using `terraform import`, import Amazon Bedrock Guardrail using using a comma-delimited string of `guardrail_id` and `version`. For example:

```console
% terraform import aws_bedrock_guardrail.example guardrail-id-12345678,DRAFT
```
