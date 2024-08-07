---
subcategory: "Amazon Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_guardrail"
description: |-
  Terraform resource for managing an AWS Amazon Bedrock Guardrail.
---
<!---
TIP: A few guiding principles for writing documentation:
1. Use simple language while avoiding jargon and figures of speech.
2. Focus on brevity and clarity to keep a reader's attention.
3. Use active voice and present tense whenever you can.
4. Document your feature as it exists now; do not mention the future or past if you can help it.
5. Use accessible and inclusive language.
--->`
# Resource: aws_bedrock_guardrail

Terraform resource for managing an AWS Amazon Bedrock Guardrail.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_guardrail" "test" {
  name                      = "test"
  blocked_input_messaging   = "test"
  blocked_outputs_messaging = "test"
  description               = "test"
  content_policy_config {
    filters_config {
      input_strength  = "MEDIUM"
      output_strength = "MEDIUM"
      type            = "HATE"
    }
  }
  sensitive_information_policy_config {
    pii_entities_config {
      action = "BLOCK"
      type   = "NAME"
    }

    regexes_config {
      action      = "BLOCK"
      description = "example regex"
      name        = "regex_example"
      pattern     = "^\\d{3}-\\d{2}-\\d{4}$"
    }
  }
  topic_policy_config {
    topics_config {
      name       = "investment_topic"
      examples   = ["Where should I invest my money ?"]
      type       = "DENY"
      definition = "Investment advice refers to inquiries, guidance, or recommendations regarding the management or allocation of funds or assets with the goal of generating returns ."
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
  tags = {
    "Modified By" = "terraform"
  }
}
```

## Argument Reference

The following arguments are required:

* `blocked_input_messaging` - (Required) Messaging for when violations are detected in text.
* `blocked_outputs_messaging` - (Required) Messaging for when violations are detected in text.
* `name` - (Required) Name of the guardrail.

The following arguments are optional:

* `content_policy_config` - (Optional) Content policy config for a guardrail. See [Content Policy Config](#content-policy-config) for more information.
* `contextual_grounding_policy_config` - (Optional) Contextual grounding policy config for a guardrail.See [Contextual Grounding Policy Config](#contextual-grounding-policy-config) for more information.
* `description` (Optional) Description of the guardrail or its version
* `kms_key_arn` (Optional) The KMS key with which the guardrail was encrypted at rest
* `sensitive_information_policy_config` (Optional) Sensitive information policy config for a guardrail. See [Sensitive Information Policy Config](#sensitive-information-policy-config) for more information.
* `tags` (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.
* `topic_policy_config` (Optional) Topic policy config for a guardrail. See [Topic Policy Config](#topic-policyconfig) for more information.
* `word_policy_config` (Optional) Word policy config for a guardrail. See [Word Policy Config](#word-policy-config) for more information.

### Content Policy Config

The `content_policy_config` configuration block supports the following arguments:
* `filters_config` - (Optional) List of content filter configs in content policy. See [Filters Config](#content-filters-config) for more information.

#### Content Filters Config

The `filters_config` configuration block supports the following arguments:
* `input_strength` - (Optional) Strength for filters.
* `output_strength` - (Optional) Strength for filters.
* `type` - (Optional) Type of filter in content policy.

### Contextual Grounding Policy Config

* `filters_config` (Attributes List) List of contextual grounding filter configs.  See [Contextual Grounding Filters Config](#contextual-grounding-filters-config) for more information.

#### Contextual Grounding Filters Config

The `filters_config` configuration block supports the following arguments:
* `threshold` - (Required) The threshold for this filter.
* `type` - (Required) Type of contextual grounding filter.

### Sensitive Information Policy Config

* `pii_entities_config` (Optional) List of entities. See [PII Entities Config](#pii-entities-config) for more information.
* `regexes_config` (Optional) List of regex. See [Regexes Config](#regexes-config) for more information.

#### PII Entities Config
* `action` (Required) Options for sensitive information action.
* `type` (Required) The currently supported PII entities

#### Regexes Config
* `action` (Required) Options for sensitive information action.
* `name` (Required) The regex name.
* `pattern` (Required) The regex pattern.
* `description` (Optional) The regex description.

### Word Policy Config
* `managed_word_lists_config` (Optional) A config for the list of managed words. See [Managed Word Lists Config](#managed-word-lists-config) for more information.
* `words_config` (Optional) List of custom word configs. (see [Words Config](#words-config))

#### Managed Word Lists Config
* `type` (Required) Options for managed words.

#### Words Config
* `text` (Required) The custom word text.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `arn` - ARN of the Guardrail.
* `created_at` - Unix epoch timestamp in seconds for when the Guardrail was created.
* `id` - ID of the Guardrail.
* `status` - Status of the Bedrock Guardrail. One of `READY`, `FAILED`.
* `version` - Version of the Guardrail.


## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Amazon Bedrock Guardrail using the `example_id_arg`. For example:

```terraform
import {
  to = aws_bedrock_guardrail.example
  id = "guardrail-id-12345678:DRAFT"
}
```

Using `terraform import`, import Amazon Bedrock Guardrail using the `example_id_arg`. For example:

```console
% terraform import aws_bedrock_guardrail.example guardrail-id-12345678:DRAFT
```
