---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_user_pool_client"
description: |-
  Provides a Cognito User Pool Client resource.
---

# Resource: aws_cognito_user_pool_client

Provides a Cognito User Pool Client resource.

To manage a User Pool Client created by another service, such as when [configuring an OpenSearch Domain to use Cognito authentication](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cognito-auth.html),
use the [`aws_cognito_managed_user_pool_client` resource](cognito_managed_user_pool_client.html) instead.

## Example Usage

### Create a basic user pool client

```terraform
resource "aws_cognito_user_pool_client" "client" {
  name = "client"

  user_pool_id = aws_cognito_user_pool.pool.id
}

resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}
```

### Create a user pool client with no SRP authentication

```terraform
resource "aws_cognito_user_pool_client" "client" {
  name = "client"

  user_pool_id = aws_cognito_user_pool.pool.id

  generate_secret     = true
  explicit_auth_flows = ["ADMIN_NO_SRP_AUTH"]
}

resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}
```

### Create a user pool client with pinpoint analytics

```terraform
resource "aws_cognito_user_pool_client" "test" {
  name         = "pool_client"
  user_pool_id = aws_cognito_user_pool.test.id

  analytics_configuration {
    application_id   = aws_pinpoint_app.test.application_id
    external_id      = "some_id"
    role_arn         = aws_iam_role.test.arn
    user_data_shared = true
  }
}

resource "aws_cognito_user_pool" "test" {
  name = "pool"
}

data "aws_caller_identity" "current" {}

resource "aws_pinpoint_app" "test" {
  name = "pinpoint"
}

data "aws_iam_policy_document" "assume_role" {
  statement {
    effect = "Allow"

    principals {
      type        = "Service"
      identifiers = ["cognito-idp.amazonaws.com"]
    }

    actions = ["sts:AssumeRole"]
  }
}

resource "aws_iam_role" "test" {
  name               = "role"
  assume_role_policy = data.aws_iam_policy_document.assume_role.json
}

data "aws_iam_policy_document" "test" {
  statement {
    effect = "Allow"

    actions = [
      "mobiletargeting:UpdateEndpoint",
      "mobiletargeting:PutEvents",
    ]

    resources = ["arn:aws:mobiletargeting:*:${data.aws_caller_identity.current.account_id}:apps/${aws_pinpoint_app.test.application_id}*"]
  }
}

resource "aws_iam_role_policy" "test" {
  name   = "role_policy"
  role   = aws_iam_role.test.id
  policy = data.aws_iam_policy_document.test.json
}
```

### Create a user pool client with Cognito as the identity provider

```terraform
resource "aws_cognito_user_pool_client" "userpool_client" {
  name                                 = "client"
  user_pool_id                         = aws_cognito_user_pool.pool.id
  callback_urls                        = ["https://example.com"]
  allowed_oauth_flows_user_pool_client = true
  allowed_oauth_flows                  = ["code", "implicit"]
  allowed_oauth_scopes                 = ["email", "openid"]
  supported_identity_providers         = ["COGNITO"]
}

resource "aws_cognito_user_pool" "pool" {
  name = "pool"
}
```

## Argument Reference

The following arguments are required:

* `name` - (Required) Name of the application client.
* `user_pool_id` - (Required) User pool the client belongs to.

The following arguments are optional:

* `access_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the access token is no longer valid and cannot be used.
  By default, the unit is hours.
  The unit can be overridden by a value in `token_validity_units.access_token`.
* `allowed_oauth_flows_user_pool_client` - (Optional) Whether the client is allowed to follow the OAuth protocol when interacting with Cognito user pools.
* `allowed_oauth_flows` - (Optional) List of allowed OAuth flows (code, implicit, client_credentials).
* `allowed_oauth_scopes` - (Optional) List of allowed OAuth scopes (phone, email, openid, profile, and aws.cognito.signin.user.admin).
* `analytics_configuration` - (Optional) Configuration block for Amazon Pinpoint analytics for collecting metrics for this user pool. [Detailed below](#analytics_configuration).
* `auth_session_validity` - (Optional) Amazon Cognito creates a session token for each API request in an authentication flow. AuthSessionValidity is the duration, in minutes, of that session token. Your user pool native user must respond to each authentication challenge before the session expires. Valid values between `3` and `15`. Default value is `3`.
* `callback_urls` - (Optional) List of allowed callback URLs for the identity providers.
* `default_redirect_uri` - (Optional) Default redirect URI. Must be in the list of callback URLs.
* `enable_token_revocation` - (Optional) Enables or disables token revocation.
* `enable_propagate_additional_user_context_data` - (Optional) Activates the propagation of additional user context data.
* `explicit_auth_flows` - (Optional) List of authentication flows (ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY, USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_USER_SRP_AUTH, ALLOW_REFRESH_TOKEN_AUTH).
* `generate_secret` - (Optional) Should an application secret be generated.
* `id_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the ID token is no longer valid and cannot be used.
  By default, the unit is hours.
  The unit can be overridden by a value in `token_validity_units.id_token`.
* `logout_urls` - (Optional) List of allowed logout URLs for the identity providers.
* `prevent_user_existence_errors` - (Optional) Choose which errors and responses are returned by Cognito APIs during authentication, account confirmation, and password recovery when the user does not exist in the user pool. When set to `ENABLED` and the user does not exist, authentication returns an error indicating either the username or password was incorrect, and account confirmation and password recovery return a response indicating a code was sent to a simulated destination. When set to `LEGACY`, those APIs will return a `UserNotFoundException` exception if the user does not exist in the user pool.
* `read_attributes` - (Optional) List of user pool attributes the application client can read from.
* `refresh_token_validity` - (Optional) Time limit, between 60 minutes and 10 years, after which the refresh token is no longer valid and cannot be used.
  By default, the unit is days.
  The unit can be overridden by a value in `token_validity_units.refresh_token`.
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

### token_validity_units

Valid values for the following arguments are: `seconds`, `minutes`, `hours` or `days`.

* `access_token` - (Optional) Time unit in for the value in `access_token_validity`, defaults to `hours`.
* `id_token` - (Optional) Time unit in for the value in `id_token_validity`, defaults to `hours`.
* `refresh_token` - (Optional) Time unit in for the value in `refresh_token_validity`, defaults to `days`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `client_secret` - Client secret of the user pool client.
* `id` - ID of the user pool client.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito User Pool Clients using the `id` of the Cognito User Pool, and the `id` of the Cognito User Pool Client. For example:

```terraform
import {
  to = aws_cognito_user_pool_client.client
  id = "us-west-2_abc123/3ho4ek12345678909nh3fmhpko"
}
```

Using `terraform import`, import Cognito User Pool Clients using the `id` of the Cognito User Pool, and the `id` of the Cognito User Pool Client. For example:

```console
% terraform import aws_cognito_user_pool_client.client us-west-2_abc123/3ho4ek12345678909nh3fmhpko
```
