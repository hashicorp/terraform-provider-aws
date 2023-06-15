---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_managed_user_pool_client"
description: |-
  Manages a Cognito User Pool Client resource created by another service.
---

# Resource: aws_cognito_managed_user_pool_client

Manages a Cognito User Pool Client resource created by another service.

**This is an advanced resource** and has special caveats to be aware of when using it. Please read this document in its entirety before using this resource.

The `aws_cognito_managed_user_pool_client` resource should only be used to manage a Cognito User Pool Client created automatically by an AWS service.
For example, when [configuring an OpenSearch Domain to use Cognito authentication](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cognito-auth.html),
the OpenSearch service will create the User Pool Client on setup and delete it when no longer needed.
Therefore, the `aws_cognito_managed_user_pool_client` resource does not _create_ or _delete_ this resource, but instead "adopts" it into management.

For normal uses of a Cognito User Pool Client, use the [`aws_cognito_managed_user_pool_client` resource](cognito_user_pool_client.html) instead.

## Example Usage

```terraform
resource "aws_cognito_managed_user_pool_client" "example" {
  name_prefix  = "AmazonOpenSearchService-example"
  user_pool_id = aws_cognito_user_pool.example.id

  depends_on = [
    aws_opensearch_domain.example,
  ]
}

resource "aws_cognito_user_pool" "example" {
  name = "example"
}

resource "aws_cognito_identity_pool" "example" {
  identity_pool_name = "example"

  lifecycle {
    ignore_changes = [cognito_identity_providers]
  }
}

resource "aws_opensearch_domain" "example" {
  domain_name = "example"

  cognito_options {
    enabled          = true
    user_pool_id     = aws_cognito_user_pool.example.id
    identity_pool_id = aws_cognito_identity_pool.example.id
    role_arn         = aws_iam_role.example.arn
  }

  ebs_options {
    ebs_enabled = true
    volume_size = 10
  }

  depends_on = [
    aws_cognito_user_pool_domain.example,
    aws_iam_role_policy_attachment.example,
  ]
}

resource "aws_iam_role" "example" {
  name               = "example-role"
  path               = "/service-role/"
  assume_role_policy = data.aws_iam_policy_document.example.json
}

data "aws_iam_policy_document" "example" {
  statement {
    sid     = ""
    actions = ["sts:AssumeRole"]
    effect  = "Allow"

    principals {
      type = "Service"
      identifiers = [
        "es.${data.aws_partition.current.dns_suffix}",
      ]
    }
  }
}

resource "aws_iam_role_policy_attachment" "example" {
  role       = aws_iam_role.example.name
  policy_arn = "arn:${data.aws_partition.current.partition}:iam::aws:policy/AmazonESCognitoAccess"
}

data "aws_partition" "current" {}
```

## Argument Reference

The following arguments are required:

* `user_pool_id` - (Required) User pool the client belongs to.
* `name_pattern` - (Required, one of `name_pattern` or `name_prefix`) Regular expression that matches the name of the desired User Pool Client.
  Must match only one User Pool Client.
* `name_prefix` - (Required, one of `name_prefix` or `name_pattern`) String that matches the beginning of the name of the desired User Pool Client.
  Must match only one User Pool Client.

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

## Attributes Reference

In addition to all arguments above, the following attributes are exported:

* `client_secret` - Client secret of the user pool client.
* `id` - ID of the user pool client.
* `name` - Name of the user pool client.

## Import

Cognito User Pool Clients can be imported using the `id` of the Cognito User Pool, and the `id` of the Cognito User Pool Client, e.g.,

```
$ terraform import aws_cognito_managed_user_pool_client.client us-west-2_abc123/3ho4ek12345678909nh3fmhpko
```
