---
subcategory: "Bedrock AgentCore"
layout: "aws"
page_title: "AWS: aws_bedrockagentcore_oauth2_credential_provider"
description: |-
  Manages an AWS Bedrock AgentCore OAuth2 Credential Provider.
---

# Resource: aws_bedrockagentcore_oauth2_credential_provider

Manages an AWS Bedrock AgentCore OAuth2 Credential Provider. OAuth2 credential providers enable secure authentication with external OAuth2/OpenID Connect identity providers for agent runtimes.

## Example Usage

### GitHub OAuth Provider

```terraform
resource "aws_bedrockagentcore_oauth2_credential_provider" "github" {
  name = "github-oauth-provider"

  config {
    github {
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

  config {
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

  config {
    custom {
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

* `name` - (Required) Name of the OAuth2 credential provider.
* `config` - (Required) OAuth2 provider configuration. Must contain exactly one provider type. See [`config`](#config) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).

### `config`

The `config` block must contain exactly one of the following provider configurations:

* `custom` - (Optional) Custom OAuth2 provider configuration. See [`custom`](#custom) below.
* `github` - (Optional) GitHub OAuth provider configuration. See [`github`](#github) below.
* `google` - (Optional) Google OAuth provider configuration. See [`google`](#google) below.
* `microsoft` - (Optional) Microsoft OAuth provider configuration. See [`microsoft`](#microsoft) below.
* `salesforce` - (Optional) Salesforce OAuth provider configuration. See [`salesforce`](#salesforce) below.
* `slack` - (Optional) Slack OAuth provider configuration. See [`slack`](#slack) below.

### `custom`

The `custom` block supports the following:

**Standard Credentials (choose one pair):**

* `client_id` - (Optional) OAuth2 client ID. Cannot be used with `client_id_wo`. Must be used together with `client_secret`.
* `client_secret` - (Optional) OAuth2 client secret. Cannot be used with `client_secret_wo`. Must be used together with `client_id`.

**Write-Only Credentials (choose one pair):**

* `client_id_wo` - (Optional) Write-only OAuth2 client ID. This value is stored in Terraform state but never returned in plan outputs. Cannot be used with `client_id`. Must be used together with `client_secret_wo` and `client_credentials_wo_version`.
* `client_secret_wo` - (Optional) Write-only OAuth2 client secret. This value is stored in Terraform state but never returned in plan outputs. Cannot be used with `client_secret`. Must be used together with `client_id_wo` and `client_credentials_wo_version`.
* `client_credentials_wo_version` - (Optional) Version number for write-only credentials. Must be incremented whenever credentials are rotated. Required when using write-only credentials.

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
* `client_credentials_wo_version` - (Optional) Version number for write-only credentials. Required when using write-only credentials.

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

* `arn` - ARN of the OAuth2 credential provider.
* `client_secret_arn` - ARN of the AWS Secrets Manager secret containing the client secret.
* `vendor` - OAuth2 provider vendor type, automatically determined from the configuration block type. Possible values: `CustomOauth2`, `GithubOauth2`, `GoogleOauth2`, `Microsoft`, `SalesforceOauth2`, `SlackOauth2`.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `30m`)
* `update` - (Default `30m`)
* `delete` - (Default `30m`)

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

## Notes

### Write-Only Credentials

Write-only credentials (`client_id_wo`, `client_secret_wo`) provide enhanced security by ensuring sensitive values are never exposed in Terraform plan outputs or logs. When using write-only credentials:

1. **Version Management**: Always increment `client_credentials_wo_version` when rotating credentials
2. **State Considerations**: Write-only values are stored in Terraform state but never appear in plan outputs
3. **Import Limitations**: Write-only credentials cannot be imported and must be specified in the configuration

### Provider Type Changes

Changing the provider type (e.g., from `github` to `custom`) will force replacement of the resource, as this fundamentally changes the OAuth2 provider configuration.

### Credential Rotation

When rotating OAuth2 credentials:

- For standard credentials: Update `client_id` and `client_secret` values
- For write-only credentials: Update `client_id_wo`, `client_secret_wo`, and increment `client_credentials_wo_version`
