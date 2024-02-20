---
subcategory: "Cognito IDP (Identity Provider)"
layout: "aws"
page_title: "AWS: aws_cognito_managed_user_pool_client"
description: |-
  Use the `aws_cognito_user_pool_client` resource to manage a Cognito User Pool Client. This resource is created by another service.
---

# Resource: aws_cognito_managed_user_pool_client

Use the `aws_cognito_user_pool_client` resource to manage a Cognito User Pool Client.

**This resource is advanced** and has special caveats to consider before use. Please read this document completely before using the resource.

Use the `aws_cognito_managed_user_pool_client` resource to manage a Cognito User Pool Client that is automatically created by an AWS service. For instance, when [configuring an OpenSearch Domain to use Cognito authentication](https://docs.aws.amazon.com/opensearch-service/latest/developerguide/cognito-auth.html), the OpenSearch service creates the User Pool Client during setup and removes it when it is no longer required. As a result, the `aws_cognito_managed_user_pool_client` resource does not create or delete this resource, but instead assumes management of it.

Use the [`aws_cognito_user_pool_client`](cognito_user_pool_client.html) resource to manage Cognito User Pool Clients for normal use cases.

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

* `user_pool_id` - (Required) User pool that the client belongs to.
* `name_pattern` - (Required, one of `name_pattern` or `name_prefix`) Regular expression that matches the name of the desired User Pool Client. It must only match one User Pool Client.
* `name_prefix` - (Required, one of `name_prefix` or `name_pattern`) String that matches the beginning of the name of the desired User Pool Client. It must match only one User Pool Client.

The following arguments are optional:

* `access_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the access token is no longer valid and cannot be used. By default, the unit is hours. The unit can be overridden by a value in `token_validity_units.access_token`.
* `allowed_oauth_flows_user_pool_client` - (Optional) Whether the client is allowed to use the OAuth protocol when interacting with Cognito user pools.
* `allowed_oauth_flows` - (Optional) List of allowed OAuth flows, including code, implicit, and client_credentials.
* `allowed_oauth_scopes` - (Optional) List of allowed OAuth scopes, including phone, email, openid, profile, and aws.cognito.signin.user.admin.
* `analytics_configuration` - (Optional) Configuration block for Amazon Pinpoint analytics that collects metrics for this user pool. See [details below](#analytics_configuration).
* `auth_session_validity` - (Optional) Duration, in minutes, of the session token created by Amazon Cognito for each API request in an authentication flow. The session token must be responded to by the native user of the user pool before it expires. Valid values for `auth_session_validity` are between `3` and `15`, with a default value of `3`.
* `callback_urls` - (Optional) List of allowed callback URLs for the identity providers.
* `default_redirect_uri` - (Optional) Default redirect URI and must be included in the list of callback URLs.
* `enable_token_revocation` - (Optional) Enables or disables token revocation.
* `enable_propagate_additional_user_context_data` - (Optional) Enables the propagation of additional user context data.
* `explicit_auth_flows` - (Optional) List of authentication flows. The available options include ADMIN_NO_SRP_AUTH, CUSTOM_AUTH_FLOW_ONLY, USER_PASSWORD_AUTH, ALLOW_ADMIN_USER_PASSWORD_AUTH, ALLOW_CUSTOM_AUTH, ALLOW_USER_PASSWORD_AUTH, ALLOW_USER_SRP_AUTH, and ALLOW_REFRESH_TOKEN_AUTH.
* `generate_secret` - (Optional) Boolean flag indicating whether an application secret should be generated.
* `id_token_validity` - (Optional) Time limit, between 5 minutes and 1 day, after which the ID token is no longer valid and cannot be used. By default, the unit is hours. The unit can be overridden by a value in `token_validity_units.id_token`.
* `logout_urls` - (Optional) List of allowed logout URLs for the identity providers.
* `prevent_user_existence_errors` - (Optional) Setting determines the errors and responses returned by Cognito APIs when a user does not exist in the user pool during authentication, account confirmation, and password recovery.
* `read_attributes` - (Optional) List of user pool attributes that the application client can read from.
* `refresh_token_validity` - (Optional) Time limit, between 60 minutes and 10 years, after which the refresh token is no longer valid and cannot be used. By default, the unit is days. The unit can be overridden by a value in `token_validity_units.refresh_token`.
* `supported_identity_providers` - (Optional) List of provider names for the identity providers that are supported on this client. It uses the `provider_name` attribute of the `aws_cognito_identity_provider` resource(s), or the equivalent string(s).
* `token_validity_units` - (Optional) Configuration block for representing the validity times in units. See details below. [Detailed below](#token_validity_units).
* `write_attributes` - (Optional) List of user pool attributes that the application client can write to.

### analytics_configuration

Either `application_arn` or `application_id` is required for this configuration block.

* `application_arn` - (Optional) Application ARN for an Amazon Pinpoint application. It conflicts with `external_id` and `role_arn`.
* `application_id` - (Optional) Unique identifier for an Amazon Pinpoint application.
* `external_id` - (Optional) ID for the Analytics Configuration and conflicts with `application_arn`.
* `role_arn` - (Optional) ARN of an IAM role that authorizes Amazon Cognito to publish events to Amazon Pinpoint analytics. It conflicts with `application_arn`.
* `user_data_shared` - (Optional) If `user_data_shared` is set to `true`, Amazon Cognito will include user data in the events it publishes to Amazon Pinpoint analytics.

### token_validity_units

Valid values for the following arguments are: `seconds`, `minutes`, `hours`, or `days`.

* `access_token` - (Optional) Time unit for the value in `access_token_validity` and defaults to `hours`.
* `id_token` - (Optional) Time unit for the value in `id_token_validity`, and it defaults to `hours`.
* `refresh_token` - (Optional) Time unit for the value in `refresh_token_validity` and defaults to `days`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `client_secret` - Client secret of the user pool client.
* `id` - Unique identifier for the user pool client.
* `name` - Name of the user pool client.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import Cognito User Pool Clients using the `id` of the Cognito User Pool and the `id` of the Cognito User Pool Client. For example:

```terraform
import {
  to = aws_cognito_managed_user_pool_client.client
  id = "us-west-2_abc123/3ho4ek12345678909nh3fmhpko"
}
```

Using `terraform import`, import Cognito User Pool Clients using the `id` of the Cognito User Pool and the `id` of the Cognito User Pool Client. For example:

```console
% terraform import aws_cognito_managed_user_pool_client.client us-west-2_abc123/3ho4ek12345678909nh3fmhpko
```
