---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_api_key_credential_provider"
description: |-
  Manages an AWS Bedrock AgentCore API Key Credential Provider.
---

# Resource: aws_bedrockagentcore_api_key_credential_provider

Manages an AWS Bedrock AgentCore API Key Credential Provider. API Key credential providers enable secure authentication with external services that use API key-based authentication for agent runtimes.

-> **Note:** Write-Only argument `api_key_wo` is available to use in place of `api_key`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

## Example Usage

### Basic Usage

```terraform
resource "aws_bedrockagentcore_api_key_credential_provider" "example" {
  name    = "example-api-key-provider"
  api_key = "your-api-key-here"
}
```

### Write-Only API Key (Recommended for Production)

```terraform
resource "aws_bedrockagentcore_api_key_credential_provider" "example" {
  name               = "example-api-key-provider"
  api_key_wo         = "your-api-key-here"
  api_key_wo_version = 1
}
```

### Customer-Managed Secret

Reference an API key already stored in a customer-managed AWS Secrets Manager secret instead of having AgentCore create and manage one.

```terraform
resource "aws_bedrockagentcore_api_key_credential_provider" "example" {
  name                  = "example-api-key-provider"
  api_key_secret_source = "EXTERNAL"

  api_key_secret_config {
    secret_id = aws_secretsmanager_secret.example.id
    json_key  = "apiKey"
  }
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the API Key credential provider. Forces replacement when changed.

The following arguments are optional:

* `api_key` - (Optional) API key value. Conflicts with `api_key_wo`. This value will be visible in Terraform plan outputs and logs.
* `api_key_wo` - (Optional, Write-Only) Write-only API key value. Conflicts with `api_key`. If set, requires `api_key_wo_version` to be set.
* `api_key_wo_version` - (Optional) Required when `api_key_wo` is set. Changing this value triggers an update to `api_key_wo`.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

**Customer-Managed Secret:**

* `api_key_secret_source` - (Optional) Source of the secret backing the credential provider. Valid values are `MANAGED` (AgentCore creates and manages the secret from the supplied `api_key`) and `EXTERNAL` (the provider references a customer-managed AWS Secrets Manager secret via `api_key_secret_config`). Changing between `MANAGED` and `EXTERNAL` forces replacement of the resource.
* `api_key_secret_config` - (Optional) Reference to a customer-managed AWS Secrets Manager secret that stores the API key. Requires `api_key_secret_source = "EXTERNAL"`. [See below](#api_key_secret_config).

### api_key_secret_config

* `secret_id` - (Required) ID of the AWS Secrets Manager secret that stores the secret value.
* `json_key` - (Required) JSON key used to extract the secret value from the AWS Secrets Manager secret.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `api_key_secret_arn` - ARN of the AWS Secrets Manager secret containing the API key.
    * `secret_arn` - ARN of the secret in AWS Secrets Manager.
* `credential_provider_arn` - ARN of the API Key credential provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_api_key_credential_provider.example
  identity = {
    name = "example-api-key-provider"
  }
}

resource "aws_bedrockagentcore_api_key_credential_provider" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) API key credential provider name.

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore API Key Credential Provider using the provider name. For example:

```terraform
import {
  to = aws_bedrockagentcore_api_key_credential_provider.example
  id = "example-api-key-provider"
}
```

Using `terraform import`, import Bedrock AgentCore API Key Credential Provider using the provider name. For example:

```console
% terraform import aws_bedrockagentcore_api_key_credential_provider.example example-api-key-provider
```
