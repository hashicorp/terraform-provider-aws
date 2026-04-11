---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_enforced_guardrail_configuration"
description: |-
  Terraform resource for managing an Amazon Bedrock Enforced Guardrail Configuration.
---

# Resource: aws_bedrock_enforced_guardrail_configuration

Terraform resource for managing an Amazon Bedrock Enforced Guardrail Configuration.

This resource configures an account-level enforced guardrail that is applied across all Bedrock inference calls in an account.

~> **NOTE:** Only one enforced guardrail configuration can exist per account. Importing or creating this resource will manage the single account-level configuration.

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "example"
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Example guardrail"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}

resource "aws_bedrock_guardrail_version" "example" {
  guardrail_arn = aws_bedrock_guardrail.example.guardrail_arn
  description   = "Example version"
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  guardrail_identifier = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version    = aws_bedrock_guardrail_version.example.version
}
```

### With Selective Content Guarding

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "example"
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Example guardrail"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  guardrail_identifier = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version    = aws_bedrock_guardrail_version.example.version

  selective_content_guarding {
    messages = "COMPREHENSIVE"
    system   = "COMPREHENSIVE"
  }
}
```

### With Model Enforcement

```terraform
resource "aws_bedrock_guardrail" "example" {
  name                      = "example"
  blocked_input_messaging   = "Blocked input"
  blocked_outputs_messaging = "Blocked output"
  description               = "Example guardrail"

  word_policy_config {
    words_config {
      text = "deny"
    }
  }
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  guardrail_identifier = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version    = aws_bedrock_guardrail_version.example.version

  model_enforcement {
    included_models = ["ALL"]
    excluded_models = []
  }
}
```

## Argument Reference

The following arguments are required:

* `guardrail_identifier` - (Required) Identifier for the guardrail, can be the guardrail ID or the guardrail ARN.
* `guardrail_version` - (Required) Numerical guardrail version. Must be a published version number (e.g., `1`, `2`). `DRAFT` is not supported.

The following arguments are optional:

* `model_enforcement` - (Optional) Model-specific information for the enforced guardrail configuration. If not present, the configuration is enforced on all models. See [`model_enforcement`](#model_enforcement) below.
* `selective_content_guarding` - (Optional) Selective content guarding controls for enforced guardrails. See [`selective_content_guarding`](#selective_content_guarding) below.

### `model_enforcement`

* `excluded_models` - (Required) List of models to exclude from enforcement of the guardrail.
* `included_models` - (Required) List of models to enforce the guardrail on. Use `ALL` to enforce on all models.

### `selective_content_guarding`

* `messages` - (Optional) Selective guarding mode for user messages. Valid values: `SELECTIVE`, `COMPREHENSIVE`.
* `system` - (Optional) Selective guarding mode for system prompts. Valid values: `SELECTIVE`, `COMPREHENSIVE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `config_id` - Unique ID for the account enforced configuration.
* `created_at` - Timestamp when the configuration was created.
* `created_by` - ARN of the role used to create the configuration.
* `guardrail_arn` - ARN of the guardrail.
* `guardrail_id` - Unique ID of the guardrail.
* `id` - AWS Region.
* `owner` - Configuration owner type. Value: `ACCOUNT`.
* `updated_at` - Timestamp when the configuration was last updated.
* `updated_by` - ARN of the role used to update the configuration.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Enforced Guardrail Configuration using the `id`. For example:

```terraform
import {
  to = aws_bedrock_enforced_guardrail_configuration.example
  id = "us-east-1"
}
```

Using `terraform import`, import Bedrock Enforced Guardrail Configuration using the `id`. For example:

```console
% terraform import aws_bedrock_enforced_guardrail_configuration.example us-east-1
```
