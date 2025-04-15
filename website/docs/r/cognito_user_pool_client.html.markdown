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

* `access_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the access token is no longer valid and cannot be used. By default, the unit is hours. The unit can be overridden by a value in `token_validity_units.access_token`.
* `allowed_oauth_flows_user_pool_client` - (Optional) Whether the client is allowed to use OAuth 2.0 features. `allowed_oauth_flows_user_pool_client` must be set to `true` before you can configure the following arguments: `callback_urls`, `logout_urls`, `allowed_oauth_scopes` and `allowed_oauth_flows`.
* `allowed_oauth_flows` - (Optional) List of allowed OAuth flows, including `code`, `implicit`, and `client_credentials`. `allowed_oauth_flows_user_pool_client` must be set to `true` before you can configure this option.
* `allowed_oauth_scopes` - (Optional) List of allowed OAuth scopes, including `phone`, `email`, `openid`, `profile`, and `aws.cognito.signin.user.admin`. `allowed_oauth_flows_user_pool_client` must be set to `true` before you can configure this option.
* `analytics_configuration` - (Optional) Configuration block for Amazon Pinpoint analytics that collects metrics for this user pool. See [details below](#analytics_configuration).
* `auth_session_validity` - (Optional) Duration, in minutes, of the session token created by Amazon Cognito for each API request in an authentication flow. The session token must be responded to by the native user of the user pool before it expires. Valid values for `auth_session_validity` are between `3` and `15`, with a default value of `3`.
* `callback_urls` - (Optional) List of allowed callback URLs for the identity providers. `allowed_oauth_flows_user_pool_client` must be set to `true` before you can configure this option.
* `default_redirect_uri` - (Optional) Default redirect URI and must be included in the list of callback URLs.
* `enable_token_revocation` - (Optional) Enables or disables token revocation.
* `enable_propagate_additional_user_context_data` - (Optional) Enables the propagation of additional user context data.
* `explicit_auth_flows` - (Optional) List of authentication flows. The available options include `ADMIN_NO_SRP_AUTH`, `CUSTOM_AUTH_FLOW_ONLY`, `USER_PASSWORD_AUTH`, `ALLOW_ADMIN_USER_PASSWORD_AUTH`, `ALLOW_CUSTOM_AUTH`, `ALLOW_USER_PASSWORD_AUTH`, `ALLOW_USER_SRP_AUTH`, `ALLOW_REFRESH_TOKEN_AUTH`, and `ALLOW_USER_AUTH`.
* `generate_secret` - (Optional) Boolean flag indicating whether an application secret should be generated.
* `id_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the ID token is no longer valid and cannot be used. By default, the unit is hours. The unit can be overridden by a value in `token_validity_units.id_token`.
* `logout_urls` - (Optional) List of allowed logout URLs for the identity providers. `allowed_oauth_flows_user_pool_client` must be set to `true` before you can configure this option.
* `prevent_user_existence_errors` - (Optional) Setting determines the errors and responses returned by Cognito APIs when a user does not exist in the user pool during authentication, account confirmation, and password recovery.
* `read_attributes` - (Optional) List of user pool attributes that the application client can read from.
* `refresh_token_validity` - (Optional) Time limit, between 60 minutes and 10 years, after which the refresh token is no longer valid and cannot be used. By default, the unit is days. The unit can be overridden by a value in `token_validity_units.refresh_token`.
* `supported_identity_providers` - (Optional) List of provider names for the identity providers that are supported on this client. It uses the `provider_name` attribute of the `aws_cognito_identity_provider` resource(s), or the equivalent string(s).
* `token_validity_units` - (Optional) Configuration block for representing the validity times in units. See details below. [Detailed below](#token_validity_units).
* `write_attributes` - (Optional) List of user pool attributes that the application client can write to.

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
