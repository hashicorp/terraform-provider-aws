---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail"
description: |-
  Terraform data source for managing an AWS Bedrock Guardrail.
---

# Data Source: aws_bedrock_guardrail

Terraform data source for managing an AWS Bedrock Guardrail.

## Example Usage

### Basic Usage

```terraform
data "aws_bedrock_guardrail" "example" {
  guardrail_identifier = "abc123def"
}
```

### Lookup by ARN

```terraform
data "aws_bedrock_guardrail" "example" {
  guardrail_identifier = "arn:aws:bedrock:us-east-1:123456789012:guardrail/abc123def"
}
```

### Fetch Latest Published Version

```terraform
data "aws_bedrock_guardrail" "example" {
  guardrail_identifier = "abc123def"
  latest = true
}
```

## Argument Reference

This data source supports the following arguments:

* `guardrail_identifier` - (Required) Unique identifier of the guardrail. This can be a guardrail ID (lowercase alphanumeric) or a full guardrail ARN.
* `latest` - (Optional) When `true`, resolves the highest published numeric version and reads its configuration. Conflicts with `version`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `version` - (Optional) Version of the guardrail to read. Defaults to `DRAFT` when not specified. Conflicts with `latest`.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `arn` - ARN of the guardrail.
* `blocked_input_messaging` - Message returned when the guardrail blocks a prompt.
* `blocked_outputs_messaging` - Message returned when the guardrail blocks a model response.
* `content_policy_config` - Content policy configuration. See [`content_policy_config`](#content_policy_config).
* `contextual_grounding_policy_config` - Contextual grounding policy configuration. See [`contextual_grounding_policy_config`](#contextual_grounding_policy_config).
* `created_at` - Date and time the guardrail was created.
* `cross_region_config` - Cross-region routing configuration. See [`cross_region_config`](#cross_region_config).
* `description` - Description of the guardrail.
* `guardrail_id` - Unique identifier of the guardrail.
* `id` - ARN of the guardrail (same as `arn`).
* `kms_key_arn` - ARN of the KMS key used to encrypt the guardrail at rest.
* `name` - Name of the guardrail.
* `sensitive_information_policy_config` - Sensitive information policy configuration. See [`sensitive_information_policy_config`](#sensitive_information_policy_config).
* `status` - Status of the guardrail. One of `READY`, `FAILED`.
* `tags` - Map of tags assigned to the guardrail.
* `topic_policy_config` - Topic policy configuration. See [`topic_policy_config`](#topic_policy_config).
* `updated_at` - Date and time the guardrail was last updated.
* `version` - Version of the guardrail that was read.
* `word_policy_config` - Word policy configuration. See [`word_policy_config`](#word_policy_config).

### `content_policy_config`

* `filters_config` - Set of content filter configurations.
  * `input_action` - Action taken when harmful content is detected in input. Valid values: `BLOCK`, `NONE`.
  * `input_enabled` - Whether guardrail evaluation is enabled on input.
  * `input_modalities` - Set of input modalities. Valid values: `IMAGE`, `TEXT`.
  * `input_strength` - Filter strength for input. Valid values: `NONE`, `LOW`, `MEDIUM`, `HIGH`.
  * `output_action` - Action taken when harmful content is detected in output. Valid values: `BLOCK`, `NONE`.
  * `output_enabled` - Whether guardrail evaluation is enabled on output.
  * `output_modalities` - Set of output modalities. Valid values: `IMAGE`, `TEXT`.
  * `output_strength` - Filter strength for output. Valid values: `NONE`, `LOW`, `MEDIUM`, `HIGH`.
  * `type` - Content filter type. Valid values: `SEXUAL`, `VIOLENCE`, `HATE`, `INSULTS`, `MISCONDUCT`, `PROMPT_ATTACK`.
* `tier_config` - Content policy tier configuration.
  * `tier_name` - Tier name. Valid values: `STANDARD`, `CLASSIC`.

### `contextual_grounding_policy_config`

* `filters_config` - List of contextual grounding filter configurations.
  * `threshold` - Threshold for the filter.
  * `type` - Type of contextual grounding filter.

### `cross_region_config`

* `guardrail_profile_identifier` - ARN of the guardrail profile used for cross-region routing.

### `sensitive_information_policy_config`

* `pii_entities_config` - List of PII entity configurations.
  * `action` - Action for sensitive information. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `input_action` - Action on input. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `input_enabled` - Whether evaluation is enabled on input.
  * `output_action` - Action on output. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `output_enabled` - Whether evaluation is enabled on output.
  * `type` - PII entity type.
* `regexes_config` - List of regex configurations.
  * `action` - Action for the regex match. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `description` - Description of the regex.
  * `input_action` - Action on input. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `input_enabled` - Whether evaluation is enabled on input.
  * `name` - Name of the regex.
  * `output_action` - Action on output. Valid values: `BLOCK`, `ANONYMIZE`, `NONE`.
  * `output_enabled` - Whether evaluation is enabled on output.
  * `pattern` - Regex pattern.

### `topic_policy_config`

* `tier_config` - Topic policy tier configuration.
  * `tier_name` - Tier name. Valid values: `STANDARD`, `CLASSIC`.
* `topics_config` - List of topic configurations.
  * `definition` - Definition of the topic.
  * `examples` - List of example phrases.
  * `name` - Name of the topic.
  * `type` - Topic type. Valid values: `DENY`.

### `word_policy_config`

* `managed_word_lists_config` - List of managed word list configurations.
  * `input_action` - Action on input. Valid values: `BLOCK`, `NONE`.
  * `input_enabled` - Whether evaluation is enabled on input.
  * `output_action` - Action on output. Valid values: `BLOCK`, `NONE`.
  * `output_enabled` - Whether evaluation is enabled on output.
  * `type` - Managed word list type. Valid values: `PROFANITY`.
* `words_config` - List of custom word configurations.
  * `input_action` - Action on input. Valid values: `BLOCK`, `NONE`.
  * `input_enabled` - Whether evaluation is enabled on input.
  * `output_action` - Action on output. Valid values: `BLOCK`, `NONE`.
  * `output_enabled` - Whether evaluation is enabled on output.
  * `text` - Custom word text.
