---
subcategory: "Bedrock"
layout: "aws"
page_title: "AWS: aws_bedrock_enforced_guardrail_configuration"
description: |-
  Terraform resource for managing an Amazon Bedrock Enforced Guardrail Configuration.
---

# Resource: aws_bedrock_enforced_guardrail_configuration

Terraform resource for managing an Amazon Bedrock Enforced Guardrail Configuration.

This resource configures an enforced guardrail that is applied across all Bedrock inference calls in the configured region.

~> Enforced guardrail configuration is configured per AWS region. To avoid overwriting settings, this resource should not be defined in multiple configurations.

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
  guardrail_arn     = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version = aws_bedrock_guardrail_version.example.version
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

resource "aws_bedrock_guardrail_version" "example" {
  guardrail_arn = aws_bedrock_guardrail.example.guardrail_arn
  description   = "Example version"
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  guardrail_arn     = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version = aws_bedrock_guardrail_version.example.version

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

resource "aws_bedrock_guardrail_version" "example" {
  guardrail_arn = aws_bedrock_guardrail.example.guardrail_arn
  description   = "Example version"
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  guardrail_arn     = aws_bedrock_guardrail.example.guardrail_arn
  guardrail_version = aws_bedrock_guardrail_version.example.version

  model_enforcement {
    included_models = ["ALL"]
    excluded_models = []
  }
}
```

## Argument Reference

The following arguments are required:

* `guardrail_arn` - (Required) Guardrail ARN.
* `guardrail_version` - (Required) Numerical guardrail version. Must be a published version number (e.g., `1`, `2`). `DRAFT` is not supported.

The following arguments are optional:

* `model_enforcement` - (Optional) Model-specific information for the enforced guardrail configuration. If not present, the configuration is enforced on all models. See [`model_enforcement` Block](#model_enforcement-block) below.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `selective_content_guarding` - (Optional) Selective content guarding controls for enforced guardrails. See [`selective_content_guarding` Block](#selective_content_guarding-block) below.

### `model_enforcement` Block

* `excluded_models` - (Required) List of models to exclude from enforcement of the guardrail.
* `included_models` - (Required) List of models to enforce the guardrail on. Use `ALL` to enforce on all models.

### `selective_content_guarding` Block

* `messages` - (Optional) Selective guarding mode for user messages. Valid values: `SELECTIVE`, `COMPREHENSIVE`.
* `system` - (Optional) Selective guarding mode for system prompts. Valid values: `SELECTIVE`, `COMPREHENSIVE`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `config_id` - Unique ID for the account enforced configuration.
* `guardrail_id` - Unique ID of the guardrail.
* `owner` - Configuration owner type. Value: `ACCOUNT`.

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrock_enforced_guardrail_configuration.example
  identity = {
    config_id = "TODO"
  }
}

resource "aws_bedrock_enforced_guardrail_configuration" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `config_id` (String) Configuration ID.

#### Optional

* `account_id` (String) AWS Account where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock Enforced Guardrail Configuration using `config_id`. For example:

```terraform
import {
  to = aws_bedrock_enforced_guardrail_configuration.example
  id = "TODO"
}
```

Using `terraform import`, import Bedrock Enforced Guardrail Configuration using `config_id`. For example:

```console
% terraform import aws_bedrock_enforced_guardrail_configuration.example TODO
```
