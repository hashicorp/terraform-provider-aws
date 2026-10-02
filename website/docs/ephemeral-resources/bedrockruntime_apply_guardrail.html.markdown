---
subcategory: "Bedrock Runtime"
layout: "aws"
page_title: "AWS: aws_bedrockruntime_apply_guardrail"
description: |-
  Evaluates text content using an Amazon Bedrock guardrail without storing the content or results in Terraform state.
---

# Ephemeral: aws_bedrockruntime_apply_guardrail

Evaluates text content using the Amazon Bedrock Runtime [ApplyGuardrail API](https://docs.aws.amazon.com/bedrock/latest/APIReference/API_runtime_ApplyGuardrail.html) without storing the content or results in Terraform state. Use this ephemeral resource to validate content during provisioning or module tests.

This resource supports text content only; it does not invoke a foundation model.

Guardrail intervention returns `action = "GUARDRAIL_INTERVENED"` rather than a Terraform error. Use a lifecycle postcondition to reject content. API errors, including access denied and invalid guardrail versions, fail the operation.

~> **Note:** Terraform can evaluate this resource during both planning and applying. Each evaluation calls the API and may incur charges. Values returned by this resource are ephemeral and can only be used in [contexts that accept ephemeral values](https://developer.hashicorp.com/terraform/language/resources/ephemeral).

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "example"
  blocked_input_messaging   = "Input blocked."
  blocked_outputs_messaging = "Output blocked."

  word_policy_config {
    words_config {
      text = "exampleblockedword"
    }
  }
}

ephemeral "aws_bedrockruntime_apply_guardrail" "example" {
  guardrail_identifier = aws_bedrock_guardrail.example.guardrail_id
  guardrail_version    = "DRAFT"
  source               = "INPUT"

  content {
    text {
      text = "Hello, world."
    }
  }

  lifecycle {
    postcondition {
      condition     = self.action == "NONE"
      error_message = "Content failed guardrail validation."
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `content` - (Required) One or more text content blocks to evaluate. See [`content` Block](#content-block) below.
* `guardrail_identifier` - (Required) ID or ARN of the guardrail. Must contain between 1 and 2048 characters.
* `guardrail_version` - (Required) Guardrail version to apply. Use `DRAFT` for the working draft or a published version number. Must contain between 1 and 8 characters.
* `source` - (Required) Source of the content. Valid values are `INPUT` and `OUTPUT`.

The following arguments are optional:

* `region` - (Optional) Region where this resource is evaluated. Defaults to the provider Region.

### `content` Block

* `text` - (Required) Exactly one text content block. See [`text` Block](#text-block) below. Image content is not supported.

### `text` Block

The following arguments are required:

* `text` - (Required) Non-empty text to evaluate. Marked as sensitive.

The following arguments are optional:

* `qualifiers` - (Optional) Qualifiers identifying the role of the text for contextual grounding checks. Valid values are `grounding_source`, `query`, and `guard_content`.

## Attribute Reference

This ephemeral resource exports the following attributes in addition to the arguments above:

* `action` - Action taken by the guardrail. Compare with `NONE` in a postcondition to require content to pass validation.
* `action_reason` - Reason for the action, when returned by AWS.
* `assessments` - List of policy assessment objects returned by the guardrail. Marked as sensitive. See [`assessments` Block](#assessments-block) below.
* `guardrail_coverage` - Singleton list containing content coverage details returned by the guardrail. See [`guardrail_coverage` Block](#guardrail_coverage-block) below.
* `output` - List of output text objects returned by the guardrail, such as a configured blocking message or redacted content. This is not an echo of the original input. See [`output` Block](#output-block) below.
* `usage` - Singleton list containing policy units processed by the guardrail. See [`usage` Block](#usage-block) below.

### `assessments` Block

* `applied_guardrail_details` - Details of the applied guardrail. See [`applied_guardrail_details` Block](#applied_guardrail_details-block) below.
* `automated_reasoning_policy` - Automated reasoning findings. See [`automated_reasoning_policy` Block](#automated_reasoning_policy-block) below.
* `content_policy` - Content policy assessment. See [`content_policy` Block](#content_policy-block) below.
* `contextual_grounding_policy` - Contextual grounding assessment. See [`contextual_grounding_policy` Block](#contextual_grounding_policy-block) below.
* `invocation_metrics` - Invocation metrics. See [`invocation_metrics` Block](#invocation_metrics-block) below.
* `sensitive_information_policy` - Sensitive information assessment. See [`sensitive_information_policy` Block](#sensitive_information_policy-block) below.
* `topic_policy` - Topic policy assessment. See [`topic_policy` Block](#topic_policy-block) below.
* `word_policy` - Word policy assessment. See [`word_policy` Block](#word_policy-block) below.

### `applied_guardrail_details` Block

* `guardrail_arn` - ARN of the applied guardrail.
* `guardrail_id` - ID of the applied guardrail.
* `guardrail_origin` - List of origins describing how the guardrail was applied.
* `guardrail_ownership` - Ownership of the guardrail, including cross-account ownership.
* `guardrail_version` - Version of the applied guardrail.

### `automated_reasoning_policy` Block

* `findings` - List of automated reasoning findings. See [`findings` Block](#findings-block) below.

### `claims` Block

* `logic` - Logical representation of the claim.
* `natural_language` - Natural language representation of the claim.

### `claims_false_scenario` Block

* `statements` - Statements describing a scenario in which the claims are false. See [`statements` Block](#statements-block) below.

### `claims_true_scenario` Block

* `statements` - Statements describing a scenario in which the claims are true. See [`statements` Block](#statements-block) below.

### `content_policy` Block

* `filters` - List of content filter assessments. See [`content_policy.filters` Block](#content_policyfilters-block) below.

### `content_policy.filters` Block

* `action` - Action taken by the content filter.
* `confidence` - Confidence level of the content classification.
* `detected` - Whether the filter detected matching content.
* `filter_strength` - Strength configured for the content filter.
* `type` - Type of content filter.

### `contextual_grounding_policy` Block

* `filters` - List of contextual grounding filter assessments. See [`contextual_grounding_policy.filters` Block](#contextual_grounding_policyfilters-block) below.

### `contextual_grounding_policy.filters` Block

* `action` - Action taken by the contextual grounding filter.
* `detected` - Whether the filter detected a violation.
* `score` - Numeric score assigned by the filter.
* `threshold` - Numeric threshold configured for the filter.
* `type` - Type of contextual grounding filter.

### `contradicting_rules` Block

* `identifier` - Identifier of the contradicting rule.
* `policy_version_arn` - ARN of the policy version containing the rule.

### `custom_words` Block

* `action` - Action taken for the custom word.
* `detected` - Whether the custom word was detected.
* `match` - Text matching the custom word.

### `difference_scenarios` Block

* `statements` - Statements describing differences between translation options. See [`statements` Block](#statements-block) below.

### `findings` Block

* `impossible` - Finding with contradictory premises. See [`impossible` Block](#impossible-block) below.
* `invalid` - Finding with claims that contradict the policy. See [`invalid` Block](#invalid-block) below.
* `no_translations` - Finding indicating no translation was produced. Represented as `[{}]`, preserving the presence of this empty AWS object.
* `satisfiable` - Finding with claims that can be either true or false under the policy. See [`satisfiable` Block](#satisfiable-block) below.
* `too_complex` - Finding indicating the input was too complex. Represented as `[{}]`.
* `translation_ambiguous` - Finding with multiple possible translations. See [`translation_ambiguous` Block](#translation_ambiguous-block) below.
* `valid` - Finding with claims supported by the policy. See [`valid` Block](#valid-block) below.

### `guardrail_coverage` Block

* `images` - Singleton list containing image coverage details. See [`images` Block](#images-block) below.
* `text_characters` - Singleton list containing text character coverage details. See [`text_characters` Block](#text_characters-block) below.

### `images` Block

* `guarded` - Number of images covered by the guardrail.
* `total` - Total number of images in the request.

### `impossible` Block

* `contradicting_rules` - List of rules that contradict the premises. See [`contradicting_rules` Block](#contradicting_rules-block) above.
* `logic_warning` - Singleton list containing a logic warning. See [`logic_warning` Block](#logic_warning-block) below.
* `translation` - Singleton list containing the logical translation. See [`translation` Block](#translation-block) below.

### `invalid` Block

* `contradicting_rules` - List of rules that contradict the claims. See [`contradicting_rules` Block](#contradicting_rules-block) above.
* `logic_warning` - Singleton list containing a logic warning. See [`logic_warning` Block](#logic_warning-block) below.
* `translation` - Singleton list containing the logical translation. See [`translation` Block](#translation-block) below.

### `invocation_metrics` Block

* `guardrail_coverage` - Singleton list containing content coverage details. See [`guardrail_coverage` Block](#guardrail_coverage-block) above.
* `guardrail_processing_latency` - Guardrail processing latency in milliseconds.
* `usage` - Singleton list containing policy usage details. See [`usage` Block](#usage-block) below.

### `logic_warning` Block

* `claims` - List of claims associated with the warning. See [`claims` Block](#claims-block) above.
* `premises` - List of premises associated with the warning. See [`premises` Block](#premises-block) below.
* `type` - Type of logic warning.

### `managed_word_lists` Block

* `action` - Action taken for the managed word list match.
* `detected` - Whether a managed word list match was detected.
* `match` - Text matching an entry in the managed word list.
* `type` - Type of managed word list.

### `options` Block

* `translations` - List of possible logical translations. See [`translations` Block](#translations-block) below.

### `output` Block

* `text` - Output text.

### `pii_entities` Block

* `action` - Action taken for the sensitive information match.
* `detected` - Whether a sensitive information match was detected.
* `match` - Text matching the sensitive information entity.
* `type` - Type of sensitive information entity.

### `premises` Block

* `logic` - Logical representation of the premise.
* `natural_language` - Natural language representation of the premise.

### `regexes` Block

* `action` - Action taken for the regular expression match.
* `detected` - Whether a regular expression match was detected.
* `match` - Text matching the regular expression.
* `name` - Name of the regular expression filter.
* `regex` - Regular expression used for matching.

### `satisfiable` Block

* `claims_false_scenario` - Singleton list containing a scenario in which the claims are false. See [`claims_false_scenario` Block](#claims_false_scenario-block) above.
* `claims_true_scenario` - Singleton list containing a scenario in which the claims are true. See [`claims_true_scenario` Block](#claims_true_scenario-block) above.
* `logic_warning` - Singleton list containing a logic warning. See [`logic_warning` Block](#logic_warning-block) above.
* `translation` - Singleton list containing the logical translation. See [`translation` Block](#translation-block) below.

### `sensitive_information_policy` Block

* `pii_entities` - List of personally identifiable information matches. See [`pii_entities` Block](#pii_entities-block) above.
* `regexes` - List of regular expression matches. See [`regexes` Block](#regexes-block) above.

### `statements` Block

* `logic` - Logical representation of the statement.
* `natural_language` - Natural language representation of the statement.

### `supporting_rules` Block

* `identifier` - Identifier of the supporting rule.
* `policy_version_arn` - ARN of the policy version containing the rule.

### `text_characters` Block

* `guarded` - Number of text characters covered by the guardrail.
* `total` - Total number of text characters in the request.

### `topic_policy` Block

* `topics` - List of topic assessments. See [`topics` Block](#topics-block) below.

### `topics` Block

* `action` - Action taken for the topic.
* `detected` - Whether the topic was detected.
* `name` - Name of the topic.
* `type` - Type of topic.

### `translation` Block

* `claims` - List of translated claims. See [`claims` Block](#claims-block) above.
* `confidence` - Numeric confidence score for the translation.
* `premises` - List of translated premises. See [`premises` Block](#premises-block) above.
* `untranslated_claims` - List of claim text fragments that could not be translated. See [`untranslated_claims` Block](#untranslated_claims-block) below.
* `untranslated_premises` - List of premise text fragments that could not be translated. See [`untranslated_premises` Block](#untranslated_premises-block) below.

### `translation_ambiguous` Block

* `difference_scenarios` - List of scenarios illustrating differences between translations. See [`difference_scenarios` Block](#difference_scenarios-block) above.
* `options` - List of translation options. See [`options` Block](#options-block) above.

### `translations` Block

* `claims` - List of translated claims. See [`claims` Block](#claims-block) above.
* `confidence` - Numeric confidence score for the translation.
* `premises` - List of translated premises. See [`premises` Block](#premises-block) above.
* `untranslated_claims` - List of claim text fragments that could not be translated. See [`untranslated_claims` Block](#untranslated_claims-block) below.
* `untranslated_premises` - List of premise text fragments that could not be translated. See [`untranslated_premises` Block](#untranslated_premises-block) below.

### `untranslated_claims` Block

* `text` - Claim text that could not be translated.

### `untranslated_premises` Block

* `text` - Premise text that could not be translated.

### `usage` Block

* `automated_reasoning_policies` - Number of automated reasoning policies processed.
* `automated_reasoning_policy_units` - Text units processed by automated reasoning policies.
* `content_policy_image_units` - Image units processed by the content policy.
* `content_policy_units` - Units processed by the content policy.
* `contextual_grounding_policy_units` - Units processed by the contextual grounding policy.
* `sensitive_information_policy_free_units` - Free units processed by the sensitive information policy.
* `sensitive_information_policy_units` - Units processed by the sensitive information policy.
* `topic_policy_units` - Units processed by the topic policy.
* `word_policy_units` - Units processed by the word policy.

### `valid` Block

* `claims_true_scenario` - Singleton list containing a scenario supporting the claims. See [`claims_true_scenario` Block](#claims_true_scenario-block) above.
* `logic_warning` - Singleton list containing a logic warning. See [`logic_warning` Block](#logic_warning-block) above.
* `supporting_rules` - List of rules supporting the claims. See [`supporting_rules` Block](#supporting_rules-block) above.
* `translation` - Singleton list containing the logical translation. See [`translation` Block](#translation-block) above.

### `word_policy` Block

* `custom_words` - List of custom word matches. See [`custom_words` Block](#custom_words-block) above.
* `managed_word_lists` - List of managed word list matches. See [`managed_word_lists` Block](#managed_word_lists-block) above.
