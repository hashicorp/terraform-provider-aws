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
          issuer                      = "https://auth.company.com/realms/production"
          authorization_endpoint      = "https://auth.company.com/realms/production/protocol/openid-connect/auth"
          token_endpoint              = "https://auth.company.com/realms/production/protocol/openid-connect/token"
          response_types              = ["code", "id_token"]
          token_endpoint_auth_methods = ["client_secret_basic"]
        }
      }
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `credential_provider_vendor` - (Required) Vendor of the OAuth2 credential provider. Valid values include `CustomOauth2`, `GithubOauth2`, `GoogleOauth2`, `MicrosoftOauth2`, `SalesforceOauth2`, `SlackOauth2`, `AtlassianOauth2`, `LinkedinOauth2`, and a number of additional supported vendors (e.g. `XOauth2`, `FacebookOauth2`, `SpotifyOauth2`) configured via `included_oauth2_provider_config`. Refer to the AWS API for the full, current list.
* `name` - (Required) Name of the OAuth2 credential provider.
* `oauth2_provider_config` - (Required) OAuth2 provider configuration. Must contain exactly one provider type. See [`oauth2_provider_config` Block](#oauth2_provider_config-block) below.

The following arguments are optional:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Key-value map of resource tags. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `oauth2_provider_config` Block

Exactly one of the following provider configuration blocks must be specified:

* `atlassian_oauth2_provider_config` - (Optional) Atlassian OAuth provider configuration. See [`atlassian_oauth2_provider_config` Block](#atlassian_oauth2_provider_config-block) below.
* `custom_oauth2_provider_config` - (Optional) Custom OAuth2 provider configuration. See [`custom_oauth2_provider_config` Block](#custom_oauth2_provider_config-block) below.
* `github_oauth2_provider_config` - (Optional) GitHub OAuth provider configuration. See [`github_oauth2_provider_config` Block](#github_oauth2_provider_config-block) below.
* `google_oauth2_provider_config` - (Optional) Google OAuth provider configuration. See [`google_oauth2_provider_config` Block](#google_oauth2_provider_config-block) below.
* `included_oauth2_provider_config` - (Optional) Configuration for an included (vendor-supported) OAuth2 provider, used for the additional supported vendors. See [`included_oauth2_provider_config` Block](#included_oauth2_provider_config-block) below.
* `linkedin_oauth2_provider_config` - (Optional) LinkedIn OAuth provider configuration. See [`linkedin_oauth2_provider_config` Block](#linkedin_oauth2_provider_config-block) below.
* `microsoft_oauth2_provider_config` - (Optional) Microsoft OAuth provider configuration. See [`microsoft_oauth2_provider_config` Block](#microsoft_oauth2_provider_config-block) below.
* `salesforce_oauth2_provider_config` - (Optional) Salesforce OAuth provider configuration. See [`salesforce_oauth2_provider_config` Block](#salesforce_oauth2_provider_config-block) below.
* `slack_oauth2_provider_config` - (Optional) Slack OAuth provider configuration. See [`slack_oauth2_provider_config` Block](#slack_oauth2_provider_config-block) below.

### `atlassian_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `custom_oauth2_provider_config` Block

* `client_authentication_method` - (Optional) Client authentication method used with the token endpoint. Valid values: `CLIENT_SECRET_BASIC`, `CLIENT_SECRET_POST`, `AWS_IAM_ID_TOKEN_JWT`.
* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.
* `oauth_discovery` - (Optional) OAuth discovery configuration. See [`oauth2_provider_config.custom_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configcustom_oauth2_provider_configoauth_discovery-block) below.
* `on_behalf_of_token_exchange_config` - (Optional) On-behalf-of token exchange configuration, enabling RFC 8693 token exchange or RFC 7523 JWT authorization grant flows. See [`on_behalf_of_token_exchange_config` Block](#on_behalf_of_token_exchange_config-block) below.
* `private_endpoint` - (Optional) Default private endpoint for the custom OAuth2 provider, enabling secure connectivity through a VPC Lattice resource configuration. See [`private_endpoint` Block](#private_endpoint-block) below.
* `private_endpoint_override` - (Optional) Private endpoint overrides for the custom OAuth2 provider configuration. See [`private_endpoint_override` Block](#private_endpoint_override-block) below.
* `private_key_jwt_config` - (Optional) Private key JWT client authentication configuration used when signing client assertions. See [`private_key_jwt_config` Block](#private_key_jwt_config-block) below.

### `github_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `google_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `included_oauth2_provider_config` Block

-> **Note:** `included_oauth2_provider_config` currently supports only vendors that have fixed, AWS-known OAuth2 endpoints (for example `XOauth2`, `FacebookOauth2`, `SpotifyOauth2`), which require nothing beyond `client_id` and `client_secret`. Isolated-tenant vendors such as `OktaOauth2`, `PingOneOauth2`, and `OneLoginOauth2` require provider-specific endpoints (`issuer`, `authorization_endpoint`, `token_endpoint`) that are not yet exposed by this resource, and will fail at create time with a `Missing TokenEndpoint` error. Support for those fields is planned in a follow-up.

* `authorization_endpoint` - (Optional) OAuth2 authorization endpoint URL.
* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.
* `issuer` - (Optional) OAuth2 authorization server issuer identifier.
* `token_endpoint` - (Optional) OAuth2 token endpoint URL.

### `linkedin_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `microsoft_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.
* `tenant_id` - (Optional) Microsoft Entra (Azure AD) tenant ID. Conflicts with `tenant_id_wo`.
* `tenant_id_wo` - (Optional, Write-Only) Write-only Microsoft Entra (Azure AD) tenant ID. Conflicts with `tenant_id`. Must be used together with `tenant_id_wo_version`.
* `tenant_id_wo_version` - (Optional) Version paired with the write-only tenant ID. Increment this value to trigger an update to `tenant_id_wo`.

### `salesforce_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `slack_oauth2_provider_config` Block

* `client_credentials_wo_version` - (Optional) Version used together with the write-only credentials. Required when `client_id_wo` and `client_secret_wo` are set. Changing this value triggers an update to `client_id_wo` and `client_secret_wo`.
* `client_id` - (Optional) OAuth2 client ID. Conflicts with `client_id_wo`. Must be used together with `client_secret`.
* `client_id_wo` - (Optional, Write-Only) Write-only OAuth2 client ID. Conflicts with `client_id`. If set, requires `client_secret_wo` and `client_credentials_wo_version` to be set.
* `client_secret` - (Optional) OAuth2 client secret. Conflicts with `client_secret_wo`. Must be used together with `client_id`.
* `client_secret_config` - (Optional) Reference to an AWS Secrets Manager secret that stores the client secret. Required when `client_secret_source` is `EXTERNAL`. See [`client_secret_config` Block](#client_secret_config-block) below.
* `client_secret_source` - (Optional) Source type of the client secret. Valid values: `MANAGED` (the service manages the secret) or `EXTERNAL` (you manage the secret in AWS Secrets Manager). Use `EXTERNAL` together with `client_secret_config`.
* `client_secret_wo` - (Optional, Write-Only) Write-only OAuth2 client secret. Conflicts with `client_secret`. If set, requires `client_id_wo` and `client_credentials_wo_version` to be set.

### `client_secret_config` Block

* `json_key` - (Required) JSON key used to extract the client secret value from the Secrets Manager secret.
* `secret_id` - (Required) ID of the AWS Secrets Manager secret that stores the client secret value.

### `oauth2_provider_config.custom_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - (Optional) Manual OAuth2 authorization server metadata configuration. Cannot be used together with `discovery_url`. See [`oauth2_provider_config.custom_oauth2_provider_config.oauth_discovery.authorization_server_metadata` Block](#oauth2_provider_configcustom_oauth2_provider_configoauth_discoveryauthorization_server_metadata-block) below.
* `discovery_url` - (Optional) OpenID Connect discovery URL (e.g., `https://provider.com/.well-known/openid-configuration`). Cannot be used together with `authorization_server_metadata`.

### `oauth2_provider_config.custom_oauth2_provider_config.oauth_discovery.authorization_server_metadata` Block

* `authorization_endpoint` - (Required) OAuth2 authorization endpoint URL.
* `issuer` - (Required) OAuth2 authorization server issuer identifier.
* `response_types` - (Optional) Set of OAuth2 response types supported by the authorization server.
* `token_endpoint` - (Required) OAuth2 token endpoint URL.
* `token_endpoint_auth_methods` - (Optional) List of authentication methods supported by the token endpoint. Must contain one or two values matching `client_secret_post` or `client_secret_basic`.

### `on_behalf_of_token_exchange_config` Block

* `grant_type` - (Required) Grant type for the on-behalf-of token exchange. Valid values: `TOKEN_EXCHANGE`, `JWT_AUTHORIZATION_GRANT`.
* `token_exchange_grant_type_config` - (Optional) Configuration specific to the `TOKEN_EXCHANGE` grant type (RFC 8693). See [`token_exchange_grant_type_config` Block](#token_exchange_grant_type_config-block) below.

### `token_exchange_grant_type_config` Block

* `actor_token_content` - (Required) Content type for the actor token in the token exchange. Valid values: `NONE`, `M2M`, `AWS_IAM_ID_TOKEN_JWT`.
* `actor_token_scopes` - (Optional) Set of scopes for the actor token. Only valid when `actor_token_content` is `M2M`.

### `private_endpoint` Block

* `managed_vpc_resource` - (Optional) Service-managed VPC resource configuration. See [`managed_vpc_resource` Block](#managed_vpc_resource-block) below.
* `self_managed_lattice_resource` - (Optional) Self-managed VPC Lattice resource configuration. See [`self_managed_lattice_resource` Block](#self_managed_lattice_resource-block) below.

### `private_endpoint_override` Block

* `domain` - (Required) Domain the private endpoint override applies to.
* `private_endpoint` - (Required) Private endpoint configuration for the domain. See [`private_endpoint` Block](#private_endpoint-block) above.

### `managed_vpc_resource` Block

* `endpoint_ip_address_type` - (Required) IP address type for the endpoint. Valid values: `IPV4`, `DUALSTACK`.
* `routing_domain` - (Optional) Routing domain for the managed VPC resource.
* `security_group_ids` - (Optional) Set of up to 5 security group IDs for the managed VPC resource.
* `subnet_ids` - (Required) Set of subnet IDs for the managed VPC resource.
* `tags` - (Optional) Key-value map of tags for the managed VPC resource.
* `vpc_identifier` - (Required) Identifier of the VPC.

### `self_managed_lattice_resource` Block

* `resource_configuration_identifier` - (Required) Identifier of the VPC Lattice resource configuration.

### `private_key_jwt_config` Block

* `additional_header_claims` - (Optional) Key-value map of additional claims to include in the JWT header.
* `additional_payload_claims` - (Optional) Key-value map of additional claims to include in the JWT payload.
* `private_key_source` - (Optional) Source of the private key used to sign the JWT. See [`private_key_source` Block](#private_key_source-block) below.
* `signing_algorithm` - (Optional) Algorithm used to sign the JWT.

### `private_key_source` Block

* `kms_key_source` - (Optional) AWS KMS key source configuration for the signing key. See [`kms_key_source` Block](#kms_key_source-block) below.

### `kms_key_source` Block

* `kms_key_arn` - (Required) ARN of the AWS KMS key used to sign the JWT.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `callback_url` - Callback URL to register on the OAuth2 credential provider as an allowed callback URL. This URL is where the OAuth2 authorization server redirects users after they complete the authorization flow.
* `client_secret_arn` - ARN of the AWS Secrets Manager secret containing the client secret. See [`client_secret_arn` Block](#client_secret_arn-block) below.
* `credential_provider_arn` - ARN of the OAuth2 credential provider.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).

### `client_secret_arn` Block

* `secret_arn` - ARN of the secret in AWS Secrets Manager.

### `oauth2_provider_config.atlassian_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.atlassian_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configatlassian_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.atlassian_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.github_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.github_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configgithub_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.github_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.google_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.google_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configgoogle_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.google_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.included_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.included_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configincluded_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.included_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.linkedin_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.linkedin_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configlinkedin_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.linkedin_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.microsoft_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.microsoft_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configmicrosoft_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.microsoft_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.salesforce_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.salesforce_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configsalesforce_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.salesforce_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `oauth2_provider_config.slack_oauth2_provider_config` Block

* `oauth_discovery` - OAuth discovery configuration resolved by the service. See [`oauth2_provider_config.slack_oauth2_provider_config.oauth_discovery` Block](#oauth2_provider_configslack_oauth2_provider_configoauth_discovery-block) below.

### `oauth2_provider_config.slack_oauth2_provider_config.oauth_discovery` Block

* `authorization_server_metadata` - OAuth2 authorization server metadata resolved by the service. See [`authorization_server_metadata` Block](#authorization_server_metadata-block) below.
* `discovery_url` - OpenID Connect discovery URL resolved by the service.

### `authorization_server_metadata` Block

* `authorization_endpoint` - OAuth2 authorization endpoint URL.
* `issuer` - OAuth2 authorization server issuer identifier.
* `response_types` - Set of OAuth2 response types supported by the authorization server.
* `token_endpoint` - OAuth2 token endpoint URL.
* `token_endpoint_auth_methods` - List of authentication methods supported by the token endpoint.

## Timeouts

[Configuration options](https://developer.hashicorp.com/terraform/language/resources/syntax#operation-timeouts):

* `create` - (Default `5m`)
* `update` - (Default `5m`)
* `delete` - (Default `5m`)

## Import

-> **Note:** OAuth2 client credentials are input-only in the AgentCore API and are not returned by the read operation. On import, `client_id`, `client_secret`, `client_secret_source`, `client_secret_config`, and the write-only `client_id_wo`/`client_secret_wo`/`client_credentials_wo_version` arguments cannot be recovered from the service, so the first `terraform plan` after import shows them as additions. Run `terraform apply` once to reconcile state from your configuration; subsequent plans are clean.

In Terraform v1.12.0 and later, the [`import` block](https://developer.hashicorp.com/terraform/language/import) can be used with the `identity` attribute. For example:

```terraform
import {
  to = aws_bedrockagentcore_oauth2_credential_provider.example
  identity = {
    name = "example-oauth2-provider"
  }
}

resource "aws_bedrockagentcore_oauth2_credential_provider" "example" {
  ### Configuration omitted for brevity ###
}
```

### Identity Schema

#### Required

* `name` (String) OAuth2 credential provider name.

#### Optional

* `account_id` (String) Account ID where this resource is managed.
* `region` (String) Region where this resource is managed.

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Bedrock AgentCore OAuth2 Credential Provider using `name`. For example:

```terraform
import {
  to = aws_bedrockagentcore_oauth2_credential_provider.example
  id = "example-oauth2-provider"
}
```

Using `terraform import`, import Bedrock AgentCore OAuth2 Credential Provider using `name`. For example:

```console
% terraform import aws_bedrockagentcore_oauth2_credential_provider.example example-oauth2-provider
```
