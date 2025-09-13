---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client"
description: |-
  Provides a Cognito User Pool Client
---

# Data Source: aws_cognito_user_pool_client

Provides a Cognito User Pool Client resource.

## Example Usage

```terraform
data "aws_cognito_user_pool_client" "client" {
  client_id    = "38fjsnc484p94kpqsnet7mpld0"
  user_pool_id = "us-west-2_aaaaaaaaa"
}
```

## Argument Reference

This data source supports the following arguments:

* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `client_id` - (Required) Client Id of the user pool.
* `user_pool_id` - (Required) User pool the client belongs to.

## Attribute Reference

This data source exports the following attributes in addition to the arguments above:

* `access_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the access token is no longer valid and cannot be used. This value will be overridden if you have entered a value in `token_validity_units`.
* `allowed_oauth_flows_user_pool_client` - (Optional) Whether the client is allowed to follow the OAuth protocol when interacting with Cognito user pools.
* `allowed_oauth_flows` - (Optional) List of allowed OAuth flows (code, implicit, client_credentials).
* `allowed_oauth_scopes` - (Optional) List of allowed OAuth scopes (phone, email, openid, profile, and aws.cognito.signin.user.admin).
* `analytics_configuration` - (Optional) Configuration block for Amazon Pinpoint analytics for collecting metrics for this user pool. [Detailed below](#analytics_configuration).
* `callback_urls` - (Optional) List of allowed callback URLs for the identity providers.
* `client_secret` - Client secret of the user pool client.
* `default_redirect_uri` - (Optional) Default redirect URI. Must be in the list of callback URLs.
* `enable_token_revocation` - (Optional) Enables or disables token revocation.
* `explicit_auth_flows` - (Optional) List of authentication flows (ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY, USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_USER_SRP_AUTH, ALLOW_REFRESH_TOKEN_AUTH).
* `generate_secret` - (Optional) Should an application secret be generated.
* `id_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the ID token is no longer valid and cannot be used. This value will be overridden if you have entered a value in `token_validity_units`.
* `logout_urls` - (Optional) List of allowed logout URLs for the identity providers.
* `prevent_user_existence_errors` - (Optional) Choose which errors and responses are returned by Cognito APIs during authentication, account confirmation, and password recovery when the user does not exist in the user pool. When set to `ENABLED` and the user does not exist, authentication returns an error indicating either the username or password was incorrect, and account confirmation and password recovery return a response indicating a code was sent to a simulated destination. When set to `LEGACY`, those APIs will return a `UserNotFoundException` exception if the user does not exist in the user pool.
* `read_attributes` - (Optional) List of user pool attributes the application client can read from.
* `refresh_token_rotation` - (Optional) A block that specifies the configuration of refresh token rotation. [Detailed below](#refresh_token_rotation).
* `refresh_token_validity` - (Optional) Time limit in days refresh tokens are valid for.
* `supported_identity_providers` - (Optional) List of provider names for the identity providers that are supported on this client. Uses the `provider_name` attribute of `aws_cognito_identity_provider` resource(s), or the equivalent string(s).
* `token_validity_units` - (Optional) Configuration block for units in which the validity times are represented in. [Detailed below](#token_validity_units).
* `write_attributes` - (Optional) List of user pool attributes the application client can write to.

### analytics_configuration

Either `application_arn` or `application_id` is required.

* `application_arn` - (Optional) Application ARN for an Amazon Pinpoint application. Conflicts with `external_id` and `role_arn`.
* `application_id` - (Optional) Application ID for an Amazon Pinpoint application.
* `external_id` - (Optional) ID for the Analytics Configuration. Conflicts with `application_arn`.
* `role_arn` - (Optional) ARN of an IAM role that authorizes Amazon Cognito to publish events to Amazon Pinpoint analytics. Conflicts with `application_arn`.
* `user_data_shared` (Optional) If set to `true`, Amazon Cognito will include user data in the events it publishes to Amazon Pinpoint analytics.

### refresh_token_rotation

* `feature` - (Required) The state of refresh token rotation for the current app client. Valid values are `ENABLED` or `DISABLED`.
* `retry_grace_period_seconds` - (Optional) A period of time in seconds that the user has to use the old refresh token before it is invalidated. Valid values are between `0` and `60`.

### token_validity_units

Valid values for the following arguments are: `seconds`, `minutes`, `hours` or `days`.

* `access_token` - (Optional) Time unit in for the value in `access_token_validity`, defaults to `hours`.
* `id_token` - (Optional) Time unit in for the value in `id_token_validity`, defaults to `hours`.
* `refresh_token` - (Optional) Time unit in for the value in `refresh_token_validity`, defaults to `days`.
