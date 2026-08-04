---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_oauth2_credential_provider"
description: |-
  Manages an AWS Bedrock AgentCore OAuth2 Credential Provider.
---

# Resource: aws_bedrockagentcore_oauth2_credential_provider

Manages an AWS Bedrock AgentCore OAuth2 Credential Provider. OAuth2 credential providers enable secure authentication with external OAuth2/OpenID Connect identity providers for agent runtimes.

-> **Note:** Write-Only arguments `client_id_wo` and `client_secret_wo` are available to use in place of `client_id` and `client_secret`. Write-Only arguments are supported in HashiCorp Terraform 1.11.0 and later. [Learn more](https://developer.hashicorp.com/terraform/language/resources/ephemeral#write-only-arguments).

## Example Usage

### GitHub OAuth Provider

```terraform
resource "aws_bedrockagentcore_oauth2_credential_provider" "github" {
  name = "github-oauth-provider"

  credential_provider_vendor = "GithubOauth2"
  oauth2_provider_config {
    github_oauth2_provider_config {
      client_id     = "your-github-client-id"
      client_secret = "your-github-client-secret"
    }
  }
}
```

### Custom OAuth Provider with Discovery URL

```terraform
resource "aws_bedrockagentcore_oauth2_credential_provider" "auth0" {
  name = "auth0-oauth-provider"

  credential_provider_vendor = "CustomOauth2"
  custom_oauth2_provider_config {
    custom {
      client_id_wo                  = "auth0-client-id"
      client_secret_wo              = "auth0-client-secret"
      client_credentials_wo_version = 1

      oauth_discovery {
        discovery_url = "https://dev-company.auth0.com/.well-known/openid-configuration"
      }
    }
  }
}
```

### Custom OAuth Provider with Authorization Server Metadata

```terraform
resource "aws_bedrockagentcore_oauth2_credential_provider" "keycloak" {
  name = "keycloak-oauth-provider"

  credential_provider_vendor = "CustomOauth2"
  oauth2_provider_config {
    custom_oauth2_provider_config {
      client_id_wo                  = "keycloak-client-id"
      client_secret_wo              = "keycloak-client-secret"
      client_credentials_wo_version = 1

      oauth_discovery {
        authorization_server_metadata {
          issuer                 = "https://auth.company.com/realms/production"
          authorization_endpoint = "https://auth.company.com/realms/production/protocol/openid-connect/auth"
          token_endpoint         = "https://auth.company.com/realms/production/protocol/openid-connect/token"
          response_types         = ["code", "id_token"]
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `credential_provider_vendor` - (Required) Vendor of the OAuth2 credential provider. Valid values: `CustomOauth2`, `GithubOauth2`, `GoogleOauth2`, `Microsoft`, `SalesforceOauth2`, `SlackOauth2`.
* `name` - (Required) Name of the OAuth2 credential provider.
* `oauth2_provider_config` - (Required) OAuth2 provider configuration. Must contain exactly one provider type. See [`oauth2_provider_config`](#oauth2_provider_config) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `oauth2_provider_config`

The `oauth2_provider_config` block must contain exactly one of the following provider configurations:

* `custom_oauth2_provider_config` - (Optional) Custom OAuth2 provider configuration. See [`custom`](#custom) below.
* `github_oauth2_provider_config` - (Optional) GitHub OAuth provider configuration. See [`github`](#github-google-microsoft-salesforce-slack) below.
* `google_oauth2_provider_config` - (Optional) Google OAuth provider configuration. See [`google`](#github-google-microsoft-salesforce-slack) below.
* `microsoft_oauth2_provider_config` - (Optional) Microsoft OAuth provider configuration. See [`microsoft`](#github-google-microsoft-salesforce-slack) below.
* `salesforce_oauth2_provider_config` - (Optional) Salesforce OAuth provider configuration. See [`salesforce`](#github-google-microsoft-salesforce-slack) below.
* `slack_oauth2_provider_config` - (Optional) Slack OAuth provider configuration. See [`slack`](#github-google-microsoft-salesforce-slack) below.

### `custom`

The `custom_oauth2_provider_config` block supports the following:

**Standard Credentials (choose one pair):**

* `client_id` - (Optional) OAuth2 client ID. Cannot be used with `client_id_wo`. Must be used together with `client_secret`.
* `client_secret` - (Optional) OAuth2 client secret. Cannot be used with `client_secret_wo`. Must be used together with `client_id`.

**Write-Only Credentials (choose one pair):**

* `client_id_wo` - (Optional) Write-only OAuth2 client ID. Cannot be used with `client_id`. Must be used together with `client_secret_wo` and `client_credentials_wo_version`.
* `client_secret_wo` - (Optional) Write-only OAuth2 client secret. Cannot be used with `client_secret`. Must be used together with `client_id_wo` and `client_credentials_wo_version`.
* `client_credentials_wo_version` - (Optional) Used together with write-only credentials to trigger an update. Increment this value when an update to `client_id_wo` or `client_secret_wo` is required.

**OAuth Discovery Configuration:**

* `oauth_discovery` - (Optional) OAuth discovery configuration. See [`oauth_discovery`](#oauth_discovery) below.

### `github`, `google`, `microsoft`, `salesforce`, `slack`

These predefined provider blocks support the following:

**Standard Credentials (choose one pair):**

* `client_id` - (Optional) OAuth2 client ID. Cannot be used with `client_id_wo`. Must be used together with `client_secret`.
* `client_secret` - (Optional) OAuth2 client secret. Cannot be used with `client_secret_wo`. Must be used together with `client_id`.

**Write-Only Credentials (choose one pair):**

* `client_id_wo` - (Optional) Write-only OAuth2 client ID. Cannot be used with `client_id`. Must be used together with `client_secret_wo` and `client_credentials_wo_version`.
* `client_secret_wo` - (Optional) Write-only OAuth2 client secret. Cannot be used with `client_secret`. Must be used together with `client_id_wo` and `client_credentials_wo_version`.
* `client_credentials_wo_version` - (Optional) Used together with write-only credentials to trigger an update. Increment this value when an update to `client_id_wo` or `client_secret_wo` is required.

**Note:** These predefined providers automatically configure OAuth discovery settings based on their respective authorization servers.

### `oauth_discovery`

The `oauth_discovery` block supports exactly one of the following:

* `discovery_url` - (Optional) OpenID Connect discovery URL (e.g., `https://provider.com/.well-known/openid-configuration`). Cannot be used together with `authorization_server_metadata`.
* `authorization_server_metadata` - (Optional) Manual OAuth2 authorization server metadata configuration. Cannot be used together with `discovery_url`. See [`authorization_server_metadata`](#authorization_server_metadata) below.

### `authorization_server_metadata`

The `authorization_server_metadata` block supports the following:

* `issuer` - (Required) OAuth2 authorization server issuer identifier.
* `authorization_endpoint` - (Required) OAuth2 authorization endpoint URL.
* `token_endpoint` - (Required) OAuth2 token endpoint URL.
* `response_types` - (Optional) Set of OAuth2 response types supported by the authorization server.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `credential_provider_arn` - ARN of the OAuth2 credential provider.
* `client_secret_arn` - ARN of the AWS Secrets Manager secret containing the client secret.
    * `secret_arn` - ARN of the secret in AWS Secrets Manager.
* `tags_all` - A map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore OAuth2 Credential Provider using the provider name. For example:

```terraform
import {
  to = aws_bedrockagentcore_oauth2_credential_provider.example
  id = "oauth2-provider-name"
}
```

Using `terraform import`, import Bedrock AgentCore OAuth2 Credential Provider using the provider name. For example:

```console
% terraform import aws_bedrockagentcore_oauth2_credential_provider.example oauth2-provider-name
```
