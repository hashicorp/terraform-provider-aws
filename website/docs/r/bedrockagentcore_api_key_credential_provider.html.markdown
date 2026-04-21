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

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the API Key credential provider. Forces replacement when changed.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

**Standard API Key (choose one approach):**

* `api_key` - (Optional) API key value. Cannot be used with `api_key_wo`. This value will be visible in Terraform plan outputs and logs.

**Write-Only API Key (choose one approach):**

* `api_key_wo` - (Optional) Write-only API key value. Cannot be used with `api_key`. Must be used together with `api_key_wo_version`.
* `api_key_wo_version` - (Optional) Used together with `api_key_wo` to trigger an update. Increment this value when an update to `api_key_wo` is required.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `credential_provider_arn` - ARN of the API Key credential provider.
* `api_key_secret_arn` - ARN of the AWS Secrets Manager secret containing the API key.
    * `secret_arn` - ARN of the secret in AWS Secrets Manager.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

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
