---
subcategory: "AppSync"
layout: "aws"
page_title: "AWS: aws_appsync_api"
description: |-
  Manages an AWS AppSync Event API.
---

# Resource: aws_appsync_api

Manages an [AWS AppSync Event API](https://docs.aws.amazon.com/appsync/latest/eventapi/event-api-concepts.html#API). Event APIs enable real-time subscriptions and event-driven communication in AppSync applications.

## Example Usage

### Basic Usage

```terraform
resource "aws_appsync_api" "example" {
  name = "example-event-api"

  event_config {
    auth_provider {
      auth_type = "API_KEY"
    }

    connection_auth_mode {
      auth_type = "API_KEY"
    }

    default_publish_auth_mode {
      auth_type = "API_KEY"
    }

    default_subscribe_auth_mode {
      auth_type = "API_KEY"
    }
  }
}
```

### With Cognito Authentication

```terraform
resource "aws_cognito_user_pool" "example" {
  name = "example-user-pool"
}

resource "aws_appsync_api" "example" {
  name = "example-event-api"

  event_config {
    auth_provider {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
      cognito_config {
        user_pool_id = aws_cognito_user_pool.example.id
        aws_region   = data.aws_region.current.region
      }
    }

    connection_auth_mode {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
    }

    default_publish_auth_mode {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
    }

    default_subscribe_auth_mode {
      auth_type = "AMAZON_COGNITO_USER_POOLS"
    }
  }
}

data "aws_region" "current" {}
```

### With Lambda Authorizer

```terraform
resource "aws_appsync_api" "example" {
  name = "example-event-api"

  event_config {
    auth_provider {
      auth_type = "AWS_LAMBDA"
      lambda_authorizer_config {
        authorizer_uri                   = aws_lambda_function.example.arn
        authorizer_result_ttl_in_seconds = 300
      }
    }

    connection_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    default_publish_auth_mode {
      auth_type = "AWS_LAMBDA"
    }

    default_subscribe_auth_mode {
      auth_type = "AWS_LAMBDA"
    }
  }
}
```

## Argument Reference

The following arguments are required:

* `event_config` - (Required) Configuration for the Event API. See [`event_config` Block](#event_config-block) below.
* `name` - (Required) Name of the Event API.

The following arguments are optional:

* `owner_contact` - (Optional) Contact information for the owner of the Event API.
* `region` - (Optional) Region where this resource will be [managed](https://docs.aws.amazon.com/general/latest/gr/rande.html#regional-endpoints). Defaults to the Region set in the [provider configuration](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#aws-configuration-reference).
* `tags` - (Optional) Map of tags to assign to the resource. If configured with a provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block) present, tags with matching keys will overwrite those defined at the provider-level.

### `event_config` Block

The `event_config` block supports the following:

* `auth_provider` - (Required) List of authentication providers. See [`event_config.auth_provider` Block](#event_configauth_provider-block) below.
* `connection_auth_mode` - (Required) List of authentication modes for connections. See [`event_config.connection_auth_mode` Block](#event_configconnection_auth_mode-block) below.
* `default_publish_auth_mode` - (Required) List of default authentication modes for publishing. See [`event_config.default_publish_auth_mode` Block](#event_configdefault_publish_auth_mode-block) below.
* `default_subscribe_auth_mode` - (Required) List of default authentication modes for subscribing. See [`event_config.default_subscribe_auth_mode` Block](#event_configdefault_subscribe_auth_mode-block) below.
* `log_config` - (Optional) Logging configuration. See [`log_config` Block](#log_config-block) below.

### `event_config.auth_provider` Block

The `auth_provider` block supports the following:

* `auth_type` - (Required) Type of authentication provider. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.
* `cognito_config` - (Optional) Configuration for Cognito user pool authentication. Required when `auth_type` is `AMAZON_COGNITO_USER_POOLS`. See [`cognito_config` Block](#cognito_config-block) below.
* `lambda_authorizer_config` - (Optional) Configuration for Lambda authorization. Required when `auth_type` is `AWS_LAMBDA`. See [`lambda_authorizer_config` Block](#lambda_authorizer_config-block) below.
* `openid_connect_config` - (Optional) Configuration for OpenID Connect. Required when `auth_type` is `OPENID_CONNECT`. See [`openid_connect_config` Block](#openid_connect_config-block) below.

### `cognito_config` Block

The `cognito_config` block supports the following:

* `app_id_client_regex` - (Optional) Regular expression for matching the client ID.
* `aws_region` - (Required) AWS region where the user pool is located.
* `user_pool_id` - (Required) ID of the Cognito user pool.

### `lambda_authorizer_config` Block

The `lambda_authorizer_config` block supports the following:

* `authorizer_result_ttl_in_seconds` - (Optional) TTL in seconds for the authorization result cache.
* `authorizer_uri` - (Required) URI of the Lambda function for authorization.
* `identity_validation_expression` - (Optional) Regular expression for identity validation.

### `openid_connect_config` Block

The `openid_connect_config` block supports the following:

* `auth_ttl` - (Optional) TTL in seconds for the authentication token.
* `client_id` - (Optional) Client ID for the OpenID Connect provider.
* `iat_ttl` - (Optional) TTL in seconds for the issued at time.
* `issuer` - (Required) Issuer URL for the OpenID Connect provider.

### `event_config.connection_auth_mode` Block

The `connection_auth_mode` block supports the following:

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

### `event_config.default_publish_auth_mode` Block

The `default_publish_auth_mode` block supports the following:

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

### `event_config.default_subscribe_auth_mode` Block

The `default_subscribe_auth_mode` block supports the following:

* `auth_type` - (Required) Type of authentication. Valid values: `API_KEY`, `AWS_IAM`, `AMAZON_COGNITO_USER_POOLS`, `OPENID_CONNECT`, `AWS_LAMBDA`.

### `log_config` Block

The `log_config` block supports the following:

* `cloudwatch_logs_role_arn` - (Required) ARN of the IAM role for CloudWatch logs.
* `log_level` - (Required) Log level. Valid values: `NONE`, `ERROR`, `ALL`, `INFO`, `DEBUG`.

## Attribute Reference

This resource exports the following attributes in addition to the arguments above:

* `api_arn` - ARN of the Event API.
* `api_id` - ID of the Event API.
* `dns` - DNS configuration for the Event API.
* `tags_all` - Map of tags assigned to the resource, including those inherited from the provider [`default_tags` configuration block](https://registry.terraform.io/providers/hashicorp/aws/latest/docs#default_tags-configuration-block).
* `waf_web_acl_arn` - ARN of the associated WAF web ACL.
* `xray_enabled` - Whether X-Ray tracing is enabled for the Event API.

## Import

In Terraform v1.5.0 and later, use an [`import` block](https://developer.hashicorp.com/terraform/language/import) to import AppSync Event API using the `api_id`. For example:

```terraform
import {
  to = aws_appsync_api.example
  id = "example-api-id"
}
```

Using `terraform import`, import AppSync Event API using the `api_id`. For example:

```console
% terraform import aws_appsync_api.example example-api-id
```
